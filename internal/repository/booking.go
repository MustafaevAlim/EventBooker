package repository

import (
	"context"
	"time"

	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/model"
)

type BookingRepository interface {
	Create(ctx context.Context, b model.BookingInCreate, status string) error
	GetListBooking(ctx context.Context, req model.BookingGetRequest) ([]model.BookingWithEventDetails, error)
	UpdateStatus(ctx context.Context, status string, eventID, userID int) error
	GetOccupiedPlace(ctx context.Context, eventID int) (int, error)
	GetExpiredBooking(ctx context.Context) ([]model.BookingGetForTG, error)
	GetCountUserBooking(ctx context.Context, id int) (int, error)
	DeleteExpiredBooking(ctx context.Context) error
}

type bookingRepository struct {
	db dbInterface
}

func NewBookingRepository(db dbInterface) BookingRepository {
	return &bookingRepository{db: db}
}

func (br *bookingRepository) Create(ctx context.Context, b model.BookingInCreate, status string) error {

	query := `INSERT INTO booking (user_id, event_id,status, expires_at, created_at)
				VALUES ($1, $2, $3, $4, $5)`
	_, err := br.db.ExecContext(ctx, query, b.UserID, b.EventID, status, b.ExpiresAt, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func getQueryFromMode(mode string) string {
	var query string

	switch mode {
	case "next":
		query = `SELECT 
					b.booking_id,
					b.event_id,
					b.user_id,
					b.status,
					b.expires_at,
					b.created_at,
					e.title,
					e.event_date,
					e.event_description
					FROM booking b
					INNER JOIN events e ON e.event_id = b.event_id
					WHERE b.user_id=$1 AND b.created_at > $2 AND b.booking_id > $3
					ORDER BY b.created_at ASC, b.booking_id ASC
					LIMIT $4`
	case "prev":
		query = `SELECT 
					b.booking_id,
					b.event_id,
					b.user_id,
					b.status,
					b.expires_at,
					b.created_at,
					e.title,
					e.event_date,
					e.event_description
					FROM booking b
					INNER JOIN events e ON e.event_id = b.event_id
					WHERE b.user_id=$1 AND ((b.created_at < $2) OR ( b.created_at = $2 AND b.booking_id < $3)) 
					ORDER BY b.created_at DESC, b.booking_id DESC
					LIMIT $4`
	}
	return query
}

func (br *bookingRepository) GetListBooking(ctx context.Context, req model.BookingGetRequest) ([]model.BookingWithEventDetails, error) {
	query := getQueryFromMode(req.Mode)
	args := []any{req.UserID, req.LastCreatedAt, req.LastID, req.PageSize}

	res, err := br.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Error().Msg(err.Error())
		}
	}()

	var b []model.BookingWithEventDetails
	for res.Next() {
		var temp model.BookingWithEventDetails
		err := res.Scan(&temp.ID, &temp.EventID, &temp.UserID,
			&temp.Status, &temp.ExpiresAt, &temp.CreatedAt, &temp.EventTitle,
			&temp.EventDate, &temp.EventDescription)
		if err != nil {
			return nil, err
		}
		b = append(b, temp)
	}

	return b, nil
}

func (br *bookingRepository) UpdateStatus(ctx context.Context, status string, eventID, userID int) error {
	query := `UPDATE booking
				SET status=$1
				WHERE user_id=$2 AND event_id=$3`
	_, err := br.db.ExecContext(ctx, query, status, userID, eventID)
	if err != nil {
		return err
	}
	return nil
}

func (br *bookingRepository) GetOccupiedPlace(ctx context.Context, eventID int) (int, error) {
	query := `SELECT COUNT(*)
				FROM booking
				WHERE event_id=$1 AND status IN ('pending', 'confirmed')`
	res := br.db.QueryRowContext(ctx, query, eventID)
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

func (br *bookingRepository) GetExpiredBooking(ctx context.Context) ([]model.BookingGetForTG, error) {
	query := `SELECT 
				b.booking_id,
				u.tg_chatid,
				e.title,
				e.event_date
				FROM booking b
				INNER JOIN events e ON b.event_id = e.event_id
				INNER JOIN users u ON b.user_id = u.user_id
				WHERE expires_at < $1 AND status='pending'`
	res, err := br.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Error().Msg(err.Error())
		}
	}()

	var b []model.BookingGetForTG
	for res.Next() {
		var temp model.BookingGetForTG
		err := res.Scan(&temp.ID, &temp.TgChatID, &temp.TitleEvent, &temp.EventDate)
		if err != nil {
			return nil, err
		}
		b = append(b, temp)
	}
	return b, nil

}

func (br *bookingRepository) GetCountUserBooking(ctx context.Context, id int) (int, error) {
	query := `SELECT COUNT(*)
				FROM booking
				WHERE user_id=$1`
	res := br.db.QueryRowContext(ctx, query, id)
	if res.Err() != nil {
		return 0, res.Err()
	}

	var count int
	err := res.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err

}

func (br *bookingRepository) DeleteExpiredBooking(ctx context.Context) error {
	query := `DELETE FROM booking
				WHERE status = 'pending' AND expires_at < $1`
	_, err := br.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return err
	}
	return err
}
