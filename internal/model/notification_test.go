package model

import (
	"testing"
	"time"
)

func TestToNotificationRecord(t *testing.T) {
	// テスト用のデータを準備
	now := time.Now()
	petNameMap := map[string]string{
		"pet1": "ポチ",
	}

	tests := []struct {
		name          string
		notification  Notification
		petNameMap    map[string]string
		wantErr       bool
		expectedTitle string
		expectedType  NotificationType
	}{
		{
			name: "予約通知の正常系",
			notification: Notification{
				Type:      NotificationTypeReservation,
				CreatedAt: now,
				Data: map[string]interface{}{
					"user_id":   "user1",
					"pet_id":    "pet1",
					"date_time": now.Format(time.RFC3339),
				},
			},
			petNameMap:    petNameMap,
			wantErr:       false,
			expectedTitle: "予約が完了しました",
			expectedType:  NotificationTypeReservation,
		},
		{
			name: "共通通知の正常系",
			notification: Notification{
				Type:      NotificationTypeCommon,
				CreatedAt: now,
				Data: map[string]interface{}{
					"user_id": "user1",
				},
			},
			petNameMap:    petNameMap,
			wantErr:       false,
			expectedTitle: "新しい通知が届きました。",
			expectedType:  NotificationTypeCommon,
		},
		{
			name: "無効なデータ形式",
			notification: Notification{
				Type:      NotificationTypeReservation,
				CreatedAt: now,
				Data:      "invalid",
			},
			petNameMap: petNameMap,
			wantErr:    true,
		},
		{
			name: "存在しないペットID",
			notification: Notification{
				Type:      NotificationTypeReservation,
				CreatedAt: now,
				Data: map[string]interface{}{
					"user_id":   "user1",
					"pet_id":    "nonexistent",
					"date_time": now.Format(time.RFC3339),
				},
			},
			petNameMap: petNameMap,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.notification.ToNotificationRecord(tt.petNameMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToNotificationRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Title != tt.expectedTitle {
				t.Errorf("ToNotificationRecord() title = %v, want %v", got.Title, tt.expectedTitle)
			}
			if got.Type != tt.expectedType {
				t.Errorf("ToNotificationRecord() type = %v, want %v", got.Type, tt.expectedType)
			}
		})
	}
}

func TestNewReservationNotification(t *testing.T) {
	now := time.Now()
	event := ReservationEvent{
		UserID:    "user1",
		DateTime:  now,
		PetID:     "pet1",
		CreatedAt: now,
	}

	notification := NewReservationNotification(event)

	if notification.Type != NotificationTypeReservation {
		t.Errorf("NewReservationNotification() type = %v, want %v", notification.Type, NotificationTypeReservation)
	}

	data, ok := notification.Data.(map[string]interface{})
	if !ok {
		t.Error("NewReservationNotification() data is not a map[string]interface{}")
	}

	if data["user_id"] != event.UserID {
		t.Errorf("NewReservationNotification() user_id = %v, want %v", data["user_id"], event.UserID)
	}
	if data["pet_id"] != event.PetID {
		t.Errorf("NewReservationNotification() pet_id = %v, want %v", data["pet_id"], event.PetID)
	}
	if data["date_time"] != event.DateTime {
		t.Errorf("NewReservationNotification() date_time = %v, want %v", data["date_time"], event.DateTime)
	}
}

func TestNewReservationNotificationRecord(t *testing.T) {
	now := time.Now()
	event := ReservationEvent{
		UserID:    "user1",
		DateTime:  now,
		PetID:     "pet1",
		CreatedAt: now,
	}

	record := NewReservationNotificationRecord(event)

	if record.UserID != event.UserID {
		t.Errorf("NewReservationNotificationRecord() user_id = %v, want %v", record.UserID, event.UserID)
	}
	if record.Type != NotificationTypeReservation {
		t.Errorf("NewReservationNotificationRecord() type = %v, want %v", record.Type, NotificationTypeReservation)
	}
	if !record.CreatedAt.After(now) {
		t.Error("NewReservationNotificationRecord() created_at should be after now")
	}
}
