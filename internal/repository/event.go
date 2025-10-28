package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/model"
)

type EventRepository interface {
	Create(ctx context.Context, e model.EventInCreate) error
	GetByID(ctx context.Context, id int) (model.EventInRepo, error)
	GetListEvents(ctx context.Context, req model.EventGetRequest) ([]model.EventInRepo, error)
	GetCountEvents(ctx context.Context) (int, error)
}

type eventRepository struct {
	db dbInterface
}

func NewEventRepository(db dbInterface) EventRepository {
	return &eventRepository{db: db}
}

func (er *eventRepository) Create(ctx context.Context, e model.EventInCreate) error {

	reservationPeriod, err := time.ParseDuration(e.ReservationPeriod)
	if err != nil {
		return fmt.Errorf("invalid reservation_period format (use '30m', '1h', etc): %w", err)
	}

	query := `INSERT INTO events (title, event_description, event_date,
				event_status, total_place, reservation_period, booking_confirmation,created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = er.db.ExecContext(ctx,
		query,
		e.Title, e.Description, e.EventDate, model.EventStatusPending, e.TotalPlace, reservationPeriod, e.BookingConfimation, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (er *eventRepository) GetByID(ctx context.Context, id int) (model.EventInRepo, error) {
	query := `SELECT *
				FROM events
				WHERE event_id=$1`
	res, err := er.db.QueryContext(ctx, query, id)
	if err != nil {
		return model.EventInRepo{}, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Error().Msg(err.Error())
		}
	}()

	var record model.EventInRepo
	if res.Next() {
		err := res.Scan(&record.ID, &record.Title, &record.Description, &record.EventDate, &record.Status,
			&record.TotalPlace, &record.ReservationPeriod, &record.BookingConfimation, &record.CreatedAt)
		if err != nil {
			return model.EventInRepo{}, err
		}
	}
	return record, nil
}

func (er *eventRepository) GetListEvents(ctx context.Context, req model.EventGetRequest) ([]model.EventInRepo, error) {
	var query string
	switch req.Mode {
	case "next":
		query = `SELECT *
					FROM events
					WHERE created_at > $1 AND event_id > $2
					ORDER BY created_at ASC, event_id ASC
					LIMIT $3`
	case "prev":
		query = `SELECT *
					FROM events
					WHERE (created_at < $1) OR (created_at = $1 AND event_id < $2)
					ORDER BY created_at DESC, event_id DESC
					LIMIT $3`
	}
	args := []any{req.LastCreatedAt, req.LastID, req.PageSize}

	res, err := er.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Error().Msg(err.Error())
		}
	}()

	var e []model.EventInRepo
	for res.Next() {
		var temp model.EventInRepo
		err := res.Scan(&temp.ID, &temp.Title, &temp.Description, &temp.EventDate, &temp.Status,
			&temp.TotalPlace, &temp.ReservationPeriod, &temp.BookingConfimation, &temp.CreatedAt)
		if err != nil {
			return nil, err
		}
		e = append(e, temp)
	}
	return e, nil
}

func (er *eventRepository) GetCountEvents(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*)
				FROM events`
	res := er.db.QueryRowContext(ctx, query)
	if res.Err() != nil {
		return 0, res.Err()
	}

	var count int
	err := res.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
