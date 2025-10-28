package service

import "errors"

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyPassed = errors.New("event already passed")
	ErrEmptyTitle         = errors.New("event title cannot be empty")
	ErrInvalidTotalPlace  = errors.New("total places must be positive")
	ErrInvalidEventDate   = errors.New("invalid event date")

	ErrBookingNotFound    = errors.New("booking not found")
	ErrBookingNotRequired = errors.New("booking not required")
	ErrNoSeatsAvailable   = errors.New("no seats available")

	ErrUserNotFound = errors.New("user not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)
