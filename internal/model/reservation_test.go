package model

import (
	"testing"
	"time"
)

func TestReservationEvent(t *testing.T) {
	now := time.Now()
	event := ReservationEvent{
		UserID:    "user1",
		DateTime:  now,
		PetID:     "pet1",
		CreatedAt: now,
	}

	if event.UserID != "user1" {
		t.Errorf("ReservationEvent.UserID = %v, want %v", event.UserID, "user1")
	}
	if event.DateTime != now {
		t.Errorf("ReservationEvent.DateTime = %v, want %v", event.DateTime, now)
	}
	if event.PetID != "pet1" {
		t.Errorf("ReservationEvent.PetID = %v, want %v", event.PetID, "pet1")
	}
	if event.CreatedAt != now {
		t.Errorf("ReservationEvent.CreatedAt = %v, want %v", event.CreatedAt, now)
	}
}

func TestReservation(t *testing.T) {
	now := time.Now()
	reservation := Reservation{
		ID:                  1,
		UserID:              "user1",
		UserName:            "テスト太郎",
		Email:               "test@example.com",
		ReservationDateTime: now,
		PetID:               "pet1",
		CreatedAt:           now,
		UpdatedAt:           now,
		Status:              "pending",
	}

	if reservation.ID != 1 {
		t.Errorf("Reservation.ID = %v, want %v", reservation.ID, 1)
	}
	if reservation.UserID != "user1" {
		t.Errorf("Reservation.UserID = %v, want %v", reservation.UserID, "user1")
	}
	if reservation.UserName != "テスト太郎" {
		t.Errorf("Reservation.UserName = %v, want %v", reservation.UserName, "テスト太郎")
	}
	if reservation.Email != "test@example.com" {
		t.Errorf("Reservation.Email = %v, want %v", reservation.Email, "test@example.com")
	}
	if reservation.ReservationDateTime != now {
		t.Errorf("Reservation.ReservationDateTime = %v, want %v", reservation.ReservationDateTime, now)
	}
	if reservation.PetID != "pet1" {
		t.Errorf("Reservation.PetID = %v, want %v", reservation.PetID, "pet1")
	}
	if reservation.CreatedAt != now {
		t.Errorf("Reservation.CreatedAt = %v, want %v", reservation.CreatedAt, now)
	}
	if reservation.UpdatedAt != now {
		t.Errorf("Reservation.UpdatedAt = %v, want %v", reservation.UpdatedAt, now)
	}
	if reservation.Status != "pending" {
		t.Errorf("Reservation.Status = %v, want %v", reservation.Status, "pending")
	}
}
