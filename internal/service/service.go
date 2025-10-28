package service

import (
	"EventBooker/internal/repository"
)

type Services struct {
	Event   EventService
	Booking BookingService
	User    UserService
}

func NewServices(s *repository.Storage) *Services {
	return &Services{
		Event:   NewEventService(s),
		Booking: NewBookingService(s),
		User:    NewUserService(s),
	}
}
