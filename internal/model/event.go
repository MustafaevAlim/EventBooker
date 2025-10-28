package model

import "time"

var (
	EventStatusPending  = "pending"
	EventSratusCanceled = "canceled"
	EventStatusExpired  = "expired"
)

type EventInResponse struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	EventDate          time.Time `json:"event_date"`
	TotalPlace         int       `json:"total_place"`
	OccupiedPlace      int       `json:"occupied_place"`
	EventStatus        string    `json:"event_status"`
	ReservationPeriod  string    `json:"reservation_period"`
	BookingConfimation bool      `json:"booking_confirmation"`
	CreatedAt          time.Time `json:"created_at"`
}

type EventInCreate struct {
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	EventDate          time.Time `json:"event_date"`
	TotalPlace         int       `json:"total_place"`
	ReservationPeriod  string    `json:"reservation_period"`
	BookingConfimation bool      `json:"booking_confirmation"`
}

type EventInRepo struct {
	ID                 int
	Title              string
	Description        string
	EventDate          time.Time
	Status             string
	TotalPlace         int
	ReservationPeriod  time.Duration
	BookingConfimation bool
	CreatedAt          time.Time
}

type EventGetRequest struct {
	LastCreatedAt time.Time
	LastID        int
	Mode          string
	PageSize      int
}
