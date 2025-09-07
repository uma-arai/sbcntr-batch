package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/uma-arai/sbcntr-batch/internal/common/config"
	"github.com/uma-arai/sbcntr-batch/internal/common/utils"
	"github.com/uma-arai/sbcntr-batch/internal/model"
	"github.com/uma-arai/sbcntr-batch/internal/service/batch"
)

const (
	projectName = "sbcntr-batch"
)

func main() {
	// コマンドライン引数のパース
	flag.Parse()
	timeout := flag.Duration("timeout", 5*time.Minute, "バッチ処理のタイムアウト時間")

	// 最後の引数として渡されたタスクトークンを取得
	// ENV=LOCALの場合はタスクトークンを取得しない
	taskToken := "DUMMY_TASK_TOKEN"
	if os.Getenv("ENV") != "LOCAL" {
		taskToken = flag.Arg(len(flag.Args()) - 1)
		if taskToken == "" {
			log.Fatalf("Task token is required")
		}
	}

	// 設定の読み込み
	cfg, err := config.LoadConfig(taskToken)
	if err != nil {
		log.Fatalf("Failed to load config: %v\nStack trace:\n%s", err, debug.Stack())
	}

	// X-Ray設定
	if cfg.EnableTracing {
		if err := xray.Configure(xray.Config{
			DaemonAddr:     "127.0.0.1:2000", // X-Rayデーモンのアドレス
			ServiceVersion: "1.0.0",
		}); err != nil {
			log.Printf("Failed to configure X-Ray: %v", err)
			// X-Ray設定失敗時はデフォルトの設定を使用
			if configErr := xray.Configure(xray.Config{}); configErr != nil {
				log.Fatalf("Failed to configure default X-Ray settings: %v", configErr)
			}
		}
		os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")
	}

	// 通知バッチサービスを作成
	service, err := batch.NewNotificationBatchService(cfg)
	if err != nil {
		log.Fatalf("Failed to create notification batch service: %v", err)
	}
	defer service.Close()

	// コンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// X-Rayセグメントの作成
	if cfg.EnableTracing {
		var seg *xray.Segment
		ctx, seg = xray.BeginSegment(ctx, projectName)
		defer seg.Close(nil)

		// セグメントにメタデータを追加
		if err := seg.AddMetadata("task_token", taskToken); err != nil {
			log.Printf("Failed to add task_token metadata: %v", err)
		}
		if err := seg.AddMetadata("timeout", timeout.String()); err != nil {
			log.Printf("Failed to add timeout metadata: %v", err)
		}
	}

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// バッチ処理の実行
	errChan := make(chan error, 1)
	go func() {
		// タスクトークンから通知データを生成
		notifications, err := generateNotificationsFromTaskToken(taskToken)
		if err != nil {
			log.Printf("Failed to generate notifications: %v", err)
			cancel()
			return
		}

		service.SetArgs(notifications)

		errChan <- utils.RunWithTimeout(ctx, *timeout, service.Run)
	}()

	// シグナルを待機
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case err := <-errChan:
		if err != nil {
			log.Printf("Batch process failed: %v\nStack trace:\n%s", err, debug.Stack())
			os.Exit(1)
		}
		log.Println("Batch process completed successfully")
	}
}

// generateNotificationsFromTaskToken はタスクトークンから通知データを生成します
func generateNotificationsFromTaskToken(taskToken string) ([]model.Notification, error) {
	// タスクトークンから通知データを取得する処理を実装
	// この例では、タスクトークンをJSONとして解析し、通知データを生成します
	var input struct {
		Notifications []struct {
			Type      string    `json:"type"`
			CreatedAt time.Time `json:"created_at"`
			Data      struct {
				UserID   string    `json:"user_id"`
				DateTime time.Time `json:"date_time"`
				PetID    string    `json:"pet_id"`
			} `json:"data"`
		} `json:"notifications"`
	}

	if err := json.Unmarshal([]byte(taskToken), &input); err != nil {
		return nil, fmt.Errorf("failed to parse task token: %w", err)
	}

	notifications := make([]model.Notification, len(input.Notifications))
	for i, notification := range input.Notifications {
		notifications[i] = model.NewReservationNotification(model.ReservationEvent{
			UserID:    notification.Data.UserID,
			DateTime:  notification.Data.DateTime,
			PetID:     notification.Data.PetID,
			CreatedAt: notification.CreatedAt,
		})
	}

	return notifications, nil
}
