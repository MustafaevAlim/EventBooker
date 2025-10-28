package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/model"
	"EventBooker/internal/repository"
)

type EventService interface {
	CreateEvent(ctx context.Context, e model.EventInCreate) error
	GetByID(ctx context.Context, id int) (model.EventInResponse, error)
	GetListEvents(ctx context.Context, req model.EventGetRequest) ([]model.EventInResponse, error)
	GetCountEvent(ctx context.Context) (int, error)
}

type eventService struct {
	storage *repository.Storage
}

func NewEventService(s *repository.Storage) EventService {
	return &eventService{storage: s}
}

func (es *eventService) CreateEvent(ctx context.Context, e model.EventInCreate) error {

	if err := validateCreateEvent(e); err != nil {
		return err
	}

	err := es.storage.Event.Create(ctx, e)
	if err != nil {
		zlog.Logger.Error().Msgf("service.EventService.CreateEvent error: %v", err)

		return err
	}
	return nil
}

func (es eventService) GetByID(ctx context.Context, id int) (model.EventInResponse, error) {
	e, err := es.storage.Event.GetByID(ctx, id)
	if err != nil {
		zlog.Logger.Error().Msgf("service.EventService.GetByID error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return model.EventInResponse{}, ErrEventNotFound
		}
		return model.EventInResponse{}, err
	}

	occupiedPlace, err := es.storage.Booking.GetOccupiedPlace(ctx, id)
	if err != nil {
		zlog.Logger.Error().Msgf("service.EventService.CreateEvent error: %v", err)
		return model.EventInResponse{}, err
	}

	return model.EventInResponse{
		Title:              e.Title,
		Description:        e.Description,
		EventDate:          e.EventDate,
		TotalPlace:         e.TotalPlace,
		OccupiedPlace:      occupiedPlace,
		EventStatus:        e.Status,
		BookingConfimation: e.BookingConfimation,
		ReservationPeriod:  e.ReservationPeriod.String(),
	}, nil
}

func (es eventService) GetListEvents(ctx context.Context, req model.EventGetRequest) ([]model.EventInResponse, error) {
	eventsInRepo, err := es.storage.Event.GetListEvents(ctx, req)
	if err != nil {
		zlog.Logger.Error().Msgf("service.EventService.GetListEvents error: %v", err)
		return nil, err
	}

	eventsInResponse := make([]model.EventInResponse, 0, len(eventsInRepo))

	for _, e := range eventsInRepo {
		occupiedPlace, err := es.storage.Booking.GetOccupiedPlace(ctx, e.ID)
		if err != nil {
			zlog.Logger.Error().Msgf("service.EventService.GetListEvents error: %v", err)
			return nil, err
		}
		eventsInResponse = append(eventsInResponse, model.EventInResponse{
			ID:                 e.ID,
			Title:              e.Title,
			Description:        e.Description,
			EventDate:          e.EventDate,
			TotalPlace:         e.TotalPlace,
			OccupiedPlace:      occupiedPlace,
			EventStatus:        e.Status,
			ReservationPeriod:  e.ReservationPeriod.String(),
			BookingConfimation: e.BookingConfimation,
			CreatedAt:          e.CreatedAt,
		})
	}
	return eventsInResponse, nil
}

func (es eventService) GetCountEvent(ctx context.Context) (int, error) {
	return es.storage.Event.GetCountEvents(ctx)
}

func validateCreateEvent(e model.EventInCreate) error {
	if e.Title == "" {
		return ErrEmptyTitle
	}

	if e.EventDate.Before(time.Now()) {
		return ErrInvalidEventDate
	}

	if e.TotalPlace <= 0 {
		return ErrInvalidTotalPlace
	}

	return nil

}
