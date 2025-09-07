package batch

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jmoiron/sqlx"
	"github.com/uma-arai/sbcntr-batch/internal/common/config"
	"github.com/uma-arai/sbcntr-batch/internal/model"
)

// MockNotificationRepository はテスト用のモックリポジトリです
type MockNotificationRepository struct {
	createNotificationsCalled bool
	createNotificationsError  error
	notifications             []model.NotificationRecord
}

func (m *MockNotificationRepository) CreateNotifications(ctx context.Context, records []model.NotificationRecord) error {
	m.createNotificationsCalled = true
	m.notifications = records
	return m.createNotificationsError
}

func (m *MockNotificationRepository) Create(ctx context.Context, tx *sqlx.Tx, record *model.NotificationRecord) error {
	return nil
}

func (m *MockNotificationRepository) GetByUserID(ctx context.Context, userID string) ([]model.NotificationRecord, error) {
	return nil, nil
}

func (m *MockNotificationRepository) UpdateIsRead(ctx context.Context, tx *sqlx.Tx, id int, isRead bool) error {
	return nil
}

// MockPetRepository はテスト用のモックリポジトリです
type MockPetRepository struct {
	getNameByIDCalled bool
	getNameByIDError  error
}

func (m *MockPetRepository) GetNameByID(ctx context.Context, id string) (string, error) {
	m.getNameByIDCalled = true
	return "TestPet", m.getNameByIDError
}

// newTestNotificationBatchService はテスト用のNotificationBatchServiceを作成します
func newTestNotificationBatchService(mockNotificationRepo *MockNotificationRepository, mockPetRepo *MockPetRepository) *NotificationBatchService {
	return &NotificationBatchService{
		notificationRepo: mockNotificationRepo,
		petRepo:          mockPetRepo,
		cfg:              &config.Config{},
	}
}

func TestNotificationBatchService_Run(t *testing.T) {
	// X-Rayのセグメントを設定
	ctx, seg := xray.BeginSegment(context.Background(), "TestNotificationBatchService_Run")
	defer seg.Close(nil)

	now := time.Now().UTC()
	tests := []struct {
		name          string
		notifications []model.Notification
		mockError     error
		wantErr       bool
	}{
		{
			name:          "0件の通知を正常に処理",
			notifications: []model.Notification{},
			mockError:     nil,
			wantErr:       false,
		},
		{
			name: "1件の通知を正常に処理",
			notifications: []model.Notification{
				{
					Type:      model.NotificationTypeReservation,
					Data:      map[string]interface{}{"user_id": "user1", "pet_id": "pet1", "date_time": now.Format(time.RFC3339)},
					CreatedAt: now,
				},
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name: "2件の通知を正常に処理",
			notifications: []model.Notification{
				{
					Type:      model.NotificationTypeReservation,
					Data:      map[string]interface{}{"user_id": "user1", "pet_id": "pet1", "date_time": now.Format(time.RFC3339)},
					CreatedAt: now,
				},
				{
					Type:      model.NotificationTypeReservation,
					Data:      map[string]interface{}{"user_id": "user2", "pet_id": "pet2", "date_time": now.Format(time.RFC3339)},
					CreatedAt: now,
				},
			},
			mockError: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotificationRepo := &MockNotificationRepository{
				createNotificationsError: tt.mockError,
			}
			mockPetRepo := &MockPetRepository{
				getNameByIDError: tt.mockError,
			}

			service := newTestNotificationBatchService(mockNotificationRepo, mockPetRepo)
			service.SetArgs(tt.notifications)
			err := service.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !mockNotificationRepo.createNotificationsCalled {
				t.Error("CreateNotifications was not called")
			}

			if len(mockNotificationRepo.notifications) != len(tt.notifications) {
				t.Errorf("Expected %d notifications, got %d", len(tt.notifications), len(mockNotificationRepo.notifications))
			}

			// 通知が1件以上ある場合はGetNameByIDが呼ばれているはず
			if len(tt.notifications) > 0 && !mockPetRepo.getNameByIDCalled {
				t.Error("GetNameByID was not called")
			}
		})
	}
}
