package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jmoiron/sqlx"
	"github.com/uma-arai/sbcntr-batch/internal/common/models"
	"github.com/uma-arai/sbcntr-batch/internal/model"
)

type ReservationRepository interface {
	BeginTx() (*sqlx.Tx, error)
	GetReservationsByStatus(ctx context.Context, status string) ([]models.Reservation, error)
	UpdateStatus(ctx context.Context, tx *sqlx.Tx, reservationID int64, status string) error
	CheckExistingReservation(ctx context.Context, petID string) (bool, error)
	CreateReservations(ctx context.Context, reservations []model.Reservation) error
}

type ReservationRepositoryImpl struct {
	db *DB
}

func NewReservationRepository(db *DB) *ReservationRepositoryImpl {
	return &ReservationRepositoryImpl{db: db}
}

// BeginTx starts a new transaction
func (r *ReservationRepositoryImpl) BeginTx() (*sqlx.Tx, error) {
	ctx, seg := xray.BeginSegment(context.Background(), "ReservationRepository.BeginTx")
	defer seg.Close(nil)

	return r.db.BeginTxx(ctx, nil)
}

// GetReservationsByStatus は、指定されたステータスの予約を取得します
func (r *ReservationRepositoryImpl) GetReservationsByStatus(ctx context.Context, status string) ([]models.Reservation, error) {
	ctx, seg := xray.BeginSubsegment(ctx, "ReservationRepository.GetReservationsByStatus")
	defer seg.Close(nil)

	query := `
		SELECT 
			id,
			user_id,
			user_name,
			email,
			reservation_date_time,
			pet_id,
			created_at,
			updated_at,
			status
		FROM reservations
		WHERE status = $1
		ORDER BY reservation_date_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		seg.Close(err)
		return nil, fmt.Errorf("failed to query reservations with status %s: %w", status, err)
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(
			&r.ReservationID,
			&r.UserID,
			&r.UserName,
			&r.Email,
			&r.ReservationDateTime,
			&r.PetID,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.Status,
		)
		if err != nil {
			seg.Close(err)
			return nil, fmt.Errorf("failed to scan reservation row: %w", err)
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		seg.Close(err)
		return nil, fmt.Errorf("error iterating reservation rows: %w", err)
	}

	return reservations, nil
}

// UpdateStatus は予約のステータスを更新します
func (r *ReservationRepositoryImpl) UpdateStatus(ctx context.Context, tx *sqlx.Tx, reservationID int64, status string) error {
	ctx, seg := xray.BeginSubsegment(ctx, "ReservationRepository.UpdateStatus")
	defer seg.Close(nil)

	query := `
		UPDATE reservations
		SET status = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := tx.ExecContext(ctx, query, status, time.Now(), reservationID)
	if err != nil {
		seg.Close(err)
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		seg.Close(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("no reservation found with ID %d", reservationID)
		seg.Close(err)
		return err
	}

	return nil
}

// CheckExistingReservation は、指定されたペットIDに対して予約が存在するかチェックします
func (r *ReservationRepositoryImpl) CheckExistingReservation(ctx context.Context, petID string) (bool, error) {
	ctx, seg := xray.BeginSubsegment(ctx, "ReservationRepository.CheckExistingReservation")
	defer seg.Close(nil)

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM reservations
			WHERE pet_id = $1
			AND status = 'confirmed'
			AND reservation_date_time > NOW()
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, petID).Scan(&exists)
	if err != nil {
		seg.Close(err)
		return false, fmt.Errorf("failed to check existing reservation: %w", err)
	}

	return exists, nil
}

// CreateReservations は複数の予約を作成します
func (r *ReservationRepositoryImpl) CreateReservations(ctx context.Context, reservations []model.Reservation) error {
	ctx, seg := xray.BeginSubsegment(ctx, "ReservationRepository.CreateReservations")
	defer seg.Close(nil)

	query := `
		INSERT INTO reservations (
			user_id,
			user_name,
			email,
			reservation_date_time,
			pet_id,
			status,
			created_at,
			updated_at
		) VALUES (
			:user_id,
			:user_name,
			:email,
			:reservation_date_time,
			:pet_id,
			:status,
			:created_at,
			:updated_at
		)
	`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		seg.Close(err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, reservation := range reservations {
		// PERF: bulk insertにしたほうがパフォーマンス上は望ましい
		_, err = tx.NamedExecContext(ctx, query, reservation)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("rollback failed: %v, original error: %v", rbErr, err)
			}
			seg.Close(err)
			return fmt.Errorf("failed to create reservation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		seg.Close(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
