package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/uma-arai/sbcntr-batch/internal/common/utils"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/uma-arai/sbcntr-batch/internal/common/config"
	"github.com/uma-arai/sbcntr-batch/internal/common/database"
	"github.com/uma-arai/sbcntr-batch/internal/model"
	"github.com/uma-arai/sbcntr-batch/internal/repository"
)

// ReservationBatchService は予約バッチ処理を担当します
type ReservationBatchService struct {
	args            []model.Reservation
	db              *database.DB
	reservationRepo repository.ReservationRepository
	sfnClient       *sfn.Client
	cfg             *config.Config
}

// NewReservationBatchService は新しいReservationBatchServiceを作成します
func NewReservationBatchService(cfg *config.Config, sfnClient *sfn.Client) (*ReservationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// database.DBをrepository.DBに変換
	repoDb := &repository.DB{DB: db.DB}

	return &ReservationBatchService{
		db:              db,
		reservationRepo: repository.NewReservationRepository(repoDb),
		sfnClient:       sfnClient,
		cfg:             cfg,
	}, nil
}

// Close は終了処理を行います
func (s *ReservationBatchService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SetArgs は予約バッチ処理の引数を設定します
func (s *ReservationBatchService) SetArgs(args []model.Reservation) {
	s.args = args
}

// Run は予約バッチ処理を実行します
func (s *ReservationBatchService) Run(ctx context.Context) error {
	// X-Rayセグメントの作成
	ctx, seg := xray.BeginSubsegment(ctx, "ReservationBatchService.Run")
	defer seg.Close(nil)

	startTime := time.Now()

	// バッチ処理を実行
	events, err := s.processReservationsByStatus(ctx, "pending")
	if err != nil {
		return utils.GetStackWithError(fmt.Errorf("failed to process pending reservations: %w", err))
	}

	// イベントを発行
	if err := s.sendTaskSuccess(ctx, events); err != nil {
		return utils.GetStackWithError(fmt.Errorf("failed to send task success: %w", err))
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// セグメントにメタデータを追加
	if err := seg.AddMetadata("duration", duration.String()); err != nil {
		log.Printf("Failed to add duration metadata: %v", err)
	}

	log.Printf("Reservation batch process completed successfully. Duration: %v", duration)
	return nil
}

// processReservationsByStatus は、指定されたステータスの予約を処理します
func (s *ReservationBatchService) processReservationsByStatus(ctx context.Context, status string) ([]model.ReservationEvent, error) {
	// 指定されたステータスの予約を取得
	reservations, err := s.reservationRepo.GetReservationsByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations with status %s: %w", status, err)
	}

	log.Printf("Found %d reservations with status %s", len(reservations), status)

	// 成功した予約のイベントを収集
	var events []model.ReservationEvent

	for _, reservation := range reservations {
		// トランザクション開始
		tx, err := s.reservationRepo.BeginTx()
		if err != nil {
			log.Printf("Failed to begin transaction for reservation %d: %v",
				reservation.ReservationID, err)
			continue
		}

		// 既存の予約をチェック
		exists, err := s.reservationRepo.CheckExistingReservation(ctx, reservation.PetID)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction for reservation %d: %v",
					reservation.ReservationID, rollbackErr)
			}
			log.Printf("Failed to check existing reservation for pet %s: %v",
				reservation.PetID, err)
			continue
		}

		if exists {
			// 既存の予約がある場合は、この予約をキャンセル
			if err := s.reservationRepo.UpdateStatus(ctx, tx, reservation.ReservationID, "cancelled"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v",
						reservation.ReservationID, rollbackErr)
				}
				log.Printf("Failed to update reservation status to cancelled: %v", err)
				continue
			}
		} else {
			// 既存の予約がない場合は、予約を確定
			if err := s.reservationRepo.UpdateStatus(ctx, tx, reservation.ReservationID, "confirmed"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v",
						reservation.ReservationID, rollbackErr)
				}
				log.Printf("Failed to update reservation status to confirmed: %v", err)
				continue
			}
		}

		// トランザクションをコミット
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction for reservation %d: %v",
				reservation.ReservationID, err)
			continue
		}

		// 成功/キャンセルした予約のイベントを収集
		events = append(events, model.ReservationEvent{
			UserID:    reservation.UserID,
			DateTime:  reservation.ReservationDateTime,
			PetID:     reservation.PetID,
			CreatedAt: reservation.CreatedAt,
		})
	}

	return events, nil
}

// sendTaskSuccess は、Step Functionsのタスク成功を通知し、イベントを返却します
func (s *ReservationBatchService) sendTaskSuccess(ctx context.Context, events []model.ReservationEvent) error {
	// ローカルの場合はStep Functionsの処理をスキップ
	if os.Getenv("ENV") == "LOCAL" || s.sfnClient == nil {
		log.Printf("Local environment detected. Skipping Step Functions task success notification")
		return nil
	}

	if s.sfnClient == nil {
		return fmt.Errorf("sfnClient is not initialized")
	}

	// イベントを通知形式に変換
	notifications := make([]model.Notification, len(events))
	for i, event := range events {
		notifications[i] = model.NewReservationNotification(event)
	}

	// 通知をJSONに変換
	output, err := json.Marshal(map[string]any{
		"notifications": notifications,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal notifications: %w", err)
	}

	// タスクトークンを設定から取得
	taskToken := s.cfg.SFN.TaskToken
	if taskToken == "" && os.Getenv("ENV") != "LOCAL" {
		return fmt.Errorf("SFN_TASK_TOKEN is not set in config")
	}

	// SendTaskSuccess APIを呼び出す
	input := &sfn.SendTaskSuccessInput{
		TaskToken: aws.String(taskToken),
		Output:    aws.String(string(output)),
	}

	_, err = s.sfnClient.SendTaskSuccess(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send task success: %w", err)
	}

	log.Printf("Successfully sent task success with notifications: %s", string(output))
	return nil
}
