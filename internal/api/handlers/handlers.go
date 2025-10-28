package handlers

import "EventBooker/internal/service"

type Handlers struct {
	Event   *EventHandler
	Booking *BookingHandler
	User    *UserHandler
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Event:   NewEventHandler(services.Event),
		Booking: NewBookingService(services.Booking),
		User:    NewUserHandler(services.User),
	}
}
