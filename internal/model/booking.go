package model

import "time"

var (
	StatusBookingPending   = "pending"
	StatusBookingConfirmed = "confirmed"
	StatusBookingCanceled  = "cancelled"
)

type BookingInCreate struct {
	UserID    int `json:"user_id"`
	EventID   int `json:"event_id"`
	ExpiresAt time.Time
}

type BookingWithEventDetails struct {
	BookingInRepo
	EventTitle       string
	EventDescription string
	EventDate        time.Time
}

type BookingInResponse struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	EventID          int       `json:"event_id"`
	Status           string    `json:"status"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
	EventTitle       string    `json:"event_title"`
	EventDescription string    `json:"event_description"`
	EventDate        time.Time `json:"event_date"`
}

type BookingInRepo struct {
	ID        int
	UserID    int
	EventID   int
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type BookingGetRequest struct {
	UserID        int
	Mode          string
	LastCreatedAt time.Time
	LastID        int
	PageSize      int
}

type BookingGetForTG struct {
	ID         int
	TgChatID   int64
	EventDate  time.Time
	TitleEvent string
}
