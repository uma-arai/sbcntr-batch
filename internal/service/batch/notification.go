package batch

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/uma-arai/sbcntr-batch/internal/common/config"
	"github.com/uma-arai/sbcntr-batch/internal/common/database"
	"github.com/uma-arai/sbcntr-batch/internal/model"
	"github.com/uma-arai/sbcntr-batch/internal/repository"
)

// NotificationBatchService は通知バッチ処理を担当します
type NotificationBatchService struct {
	args             []model.Notification
	db               *database.DB
	notificationRepo repository.NotificationRepository
	petRepo          repository.PetRepository
	cfg              *config.Config
}

// NewNotificationBatchService は新しいNotificationBatchServiceを作成します
func NewNotificationBatchService(cfg *config.Config) (*NotificationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// database.DBをrepository.DBに変換
	repoDb := &repository.DB{DB: db.DB}

	return &NotificationBatchService{
		db:               db,
		notificationRepo: repository.NewNotificationRepository(repoDb),
		petRepo:          repository.NewPetRepository(repoDb),
		cfg:              cfg,
	}, nil
}

// Close は終了処理を行います
func (s *NotificationBatchService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// setArgs は通知バッチ処理の引数を設定します
func (s *NotificationBatchService) SetArgs(args []model.Notification) {
	s.args = args
}

// Run は通知バッチ処理を実行します
func (s *NotificationBatchService) Run(ctx context.Context) error {
	// X-Rayセグメントの作成
	ctx, seg := xray.BeginSubsegment(ctx, "NotificationBatchService.Run")
	defer seg.Close(nil)

	notifications := s.args
	log.Printf("Starting notification batch process for %d notifications...", len(notifications))

	// セグメントにメタデータを追加
	if err := seg.AddMetadata("notification_count", len(notifications)); err != nil {
		log.Printf("Failed to add notification_count metadata: %v", err)
	}

	// 処理開始時刻を記録
	startTime := time.Now()

	// ペット名を取得
	petNameMap, err := s.getPetNameMap(ctx, notifications)
	if err != nil {
		seg.Close(err)
		return err
	}

	// 通知をレコードに変換
	records := make([]model.NotificationRecord, len(notifications))
	for i, notification := range notifications {
		record, err := notification.ToNotificationRecord(petNameMap)
		if err != nil {
			seg.Close(err)
			return err
		}
		records[i] = *record
	}

	// 通知レコードを作成
	if err := s.notificationRepo.CreateNotifications(ctx, records); err != nil {
		seg.Close(err)
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	// 処理終了時刻を記録し、実行時間を計算
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// セグメントにメタデータを追加
	if err := seg.AddMetadata("duration", duration.String()); err != nil {
		log.Printf("Failed to add duration metadata: %v", err)
	}
	if err := seg.AddMetadata("pet_count", len(petNameMap)); err != nil {
		log.Printf("Failed to add pet_count metadata: %v", err)
	}

	log.Printf("Notification batch process completed successfully. Duration: %v", duration)
	return nil
}

// 通知データに含まれる情報からペット名を取得する
// N+1とならないように先に重複がないペットIDを取得をしておく
// 1. 重複がないペットIDを取得
// 2. ペットIDからペット名を取得してMapとして保持する
func (s *NotificationBatchService) getPetNameMap(ctx context.Context, notifications []model.Notification) (map[string]string, error) {
	// X-Rayセグメントの作成
	ctx, seg := xray.BeginSubsegment(ctx, "NotificationBatchService.getPetNameMap")
	defer seg.Close(nil)

	petIDs := make([]string, 0)
	petNameMap := make(map[string]string)
	for _, notification := range notifications {
		// Dataフィールドの型をチェック
		data, ok := notification.Data.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("invalid notification data format")
			seg.Close(err)
			return nil, err
		}

		petID, ok := data["pet_id"].(string)
		if !ok {
			err := fmt.Errorf("pet_id is not a string")
			seg.Close(err)
			return nil, err
		}

		// petIDが重複している場合はスキップ
		if slices.Contains(petIDs, petID) {
			continue
		}

		petIDs = append(petIDs, petID)
	}

	// セグメントにメタデータを追加
	if err := seg.AddMetadata("unique_pet_count", len(petIDs)); err != nil {
		log.Printf("Failed to add unique_pet_count metadata: %v", err)
	}

	// ペット名を取得
	for _, petID := range petIDs {
		petName, err := s.petRepo.GetNameByID(ctx, petID)
		if err != nil {
			seg.Close(err)
			return nil, err
		}
		petNameMap[petID] = petName
	}

	return petNameMap, nil
}
