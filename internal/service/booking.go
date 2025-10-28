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

type BookingService interface {
	Book(ctx context.Context, b model.BookingInCreate) error
	Confirm(ctx context.Context, bookID, eventID, userID int) error
	GetByUserID(ctx context.Context, req model.BookingGetRequest) ([]model.BookingInResponse, error)
	GetCountUserBooking(ctx context.Context, userID int) (int, error)
	CancelBook(ctx context.Context, bookID, eventID, userID int) error
}

type bookingService struct {
	storage *repository.Storage
}

func NewBookingService(s *repository.Storage) BookingService {
	return &bookingService{storage: s}
}

func (bs *bookingService) Book(ctx context.Context, b model.BookingInCreate) error {
	return bs.storage.WithTx(ctx, func(s *repository.Storage) error {
		event, err := s.Event.GetByID(ctx, b.EventID)
		if err != nil {
			zlog.Logger.Error().Msgf("service.BookingService.Book error: %v", err)
			if errors.Is(err, sql.ErrNoRows) {
				return ErrBookingNotFound
			}
			return err
		}
		occupiedPlace, err := s.Booking.GetOccupiedPlace(ctx, b.EventID)
		if err != nil {
			zlog.Logger.Error().Msgf("service.BookingService.Book error: %v", err)
			return err
		}

		if event.EventDate.Before(time.Now()) {
			return ErrEventAlreadyPassed
		}

		if occupiedPlace >= event.TotalPlace {
			return ErrNoSeatsAvailable
		}

		b.ExpiresAt = time.Now().Add(event.ReservationPeriod)

		var status string
		if event.BookingConfimation {
			status = model.StatusBookingPending

		} else {
			status = model.StatusBookingConfirmed
		}

		err = s.Booking.Create(ctx, b, status)
		if err != nil {
			zlog.Logger.Error().Msgf("service.BookingService.Book error: %v", err)

			return err
		}
		return nil
	})
}

func (bs *bookingService) Confirm(ctx context.Context, bookID, eventID, userID int) error {
	return bs.storage.WithTx(ctx, func(s *repository.Storage) error {

		event, err := s.Event.GetByID(ctx, eventID)
		if err != nil {
			zlog.Logger.Error().Msgf("service.BookingService.Confirm error: %v", err)
			if errors.Is(err, sql.ErrNoRows) {
				return ErrBookingNotFound
			}
			return err
		}

		if !event.BookingConfimation {
			return ErrBookingNotRequired
		}

		err = bs.storage.Booking.UpdateStatus(ctx, model.StatusBookingConfirmed, bookID, eventID, userID)
		if err != nil {
			zlog.Logger.Error().Msgf("service.BookingService.Confirm error: %v", err)
			return err
		}
		return nil
	})

}

func (bs *bookingService) GetByUserID(ctx context.Context, req model.BookingGetRequest) ([]model.BookingInResponse, error) {
	bookingInRepo, err := bs.storage.Booking.GetListBooking(ctx, req)
	if err != nil {
		zlog.Logger.Error().Msgf("service.BookingService.GetByUserID error: %v", err)

		return nil, err
	}

	bookingInResponse := make([]model.BookingInResponse, 0, len(bookingInRepo))

	for _, b := range bookingInRepo {
		bookingInResponse = append(bookingInResponse, model.BookingInResponse{
			ID:               b.ID,
			UserID:           b.UserID,
			EventID:          b.EventID,
			Status:           b.Status,
			ExpiresAt:        b.ExpiresAt,
			CreatedAt:        b.CreatedAt,
			EventTitle:       b.EventTitle,
			EventDescription: b.EventDescription,
			EventDate:        b.EventDate,
		})
	}

	return bookingInResponse, nil
}

func (bs *bookingService) GetCountUserBooking(ctx context.Context, userID int) (int, error) {
	return bs.storage.Booking.GetCountUserBooking(ctx, userID)
}

func (bs *bookingService) CancelBook(ctx context.Context, bookID, eventID, userID int) error {
	return bs.storage.Booking.UpdateStatus(ctx, model.StatusBookingCanceled, bookID, eventID, userID)
}
