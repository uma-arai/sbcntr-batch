package model

import (
	"fmt"
	"time"
)

// NotificationType は通知の種類を表します
type NotificationType string

const (
	// NotificationTypeReservation は予約関連の通知を表します
	NotificationTypeReservation NotificationType = "reservation"
	// NotificationTypeCommon は共通の通知を表します
	NotificationTypeCommon NotificationType = "common"
)

// Notification はイベントIFを受け取るための定義です
// アプリケーションサービス層で利用されます
type Notification struct {
	Type      NotificationType `json:"type"`
	CreatedAt time.Time        `json:"created_at"`
	Data      interface{}      `json:"data"`
}

// NotificationRecord は通知のドメインモデルです
// データベースに永続化される通知レコードと今回は一致しています
type NotificationRecord struct {
	ID        int              `db:"id"`
	UserID    string           `db:"user_id"`
	Title     string           `db:"title"`
	Message   string           `db:"message"`
	IsRead    bool             `db:"is_read"`
	Type      NotificationType `db:"type"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}

// ToNotificationRecord は通知を通知レコードに変換します
func (n Notification) ToNotificationRecord(petNameMap map[string]string) (*NotificationRecord, error) {
	// Dataフィールドの型をチェック
	if _, ok := n.Data.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("invalid notification data format")
	}

	data := n.Data.(map[string]interface{})

	if n.Type == NotificationTypeReservation {
		petId := data["pet_id"].(string)
		petName, ok := petNameMap[petId]
		if !ok {
			return nil, fmt.Errorf("pet_id not found in petNameMap")
		}

		// date_timeフィールドの型をチェックして適切に処理
		var dateTime time.Time
		switch v := data["date_time"].(type) {
		case time.Time:
			dateTime = v
		case string:
			parsedTime, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("invalid date_time format: %v", err)
			}
			dateTime = parsedTime
		default:
			return nil, fmt.Errorf("unexpected type for date_time: %T", v)
		}

		message := fmt.Sprintf(`予約が完了しました。見学をお楽しみください。
予約日時: %s
ペット名: %s`, dateTime.Format("2006-01-02 15:04"), petName)

		return &NotificationRecord{
			UserID:    data["user_id"].(string),
			Title:     "予約が完了しました",
			Message:   message,
			IsRead:    false,
			Type:      NotificationTypeReservation,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.CreatedAt,
		}, nil
	}

	return &NotificationRecord{
		UserID:    data["user_id"].(string),
		Title:     "新しい通知が届きました。",
		Message:   "新しい通知です。",
		IsRead:    false,
		Type:      NotificationTypeCommon,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.CreatedAt,
	}, nil
}

// NewReservationNotification は予約イベントから通知を作成します
func NewReservationNotification(event ReservationEvent) Notification {
	return Notification{
		Type:      NotificationTypeReservation,
		CreatedAt: event.CreatedAt,
		Data: map[string]interface{}{
			"user_id":   event.UserID,
			"pet_id":    event.PetID,
			"date_time": event.DateTime,
		},
	}
}

// NewReservationNotificationRecord は予約イベントから通知レコードを作成します
func NewReservationNotificationRecord(event ReservationEvent) NotificationRecord {
	now := time.Now()
	return NotificationRecord{
		UserID:    event.UserID,
		Title:     "予約の更新",
		Message:   "予約のステータスが更新されました",
		IsRead:    false,
		Type:      NotificationTypeReservation,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
