package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type dbInterface interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Storage struct {
	Event   EventRepository
	Booking BookingRepository
	User    UserRepository
	db      *dbpg.DB
}

func NewStorage(db *dbpg.DB) *Storage {
	return &Storage{
		Event:   NewEventRepository(db),
		Booking: NewBookingRepository(db),
		User:    NewUserRepository(db),
		db:      db,
	}
}

func (s *Storage) WithTx(ctx context.Context, fn func(*Storage) error) error {
	if err := s.db.Master.PingContext(ctx); err != nil {
		return fmt.Errorf("database unavailable: %w", err)
	}

	tx, err := s.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txStorage := &Storage{
		Event:   NewEventRepository(tx),
		Booking: NewBookingRepository(tx),
		User:    NewUserRepository(tx),
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txStorage); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			zlog.Logger.Warn().
				Err(rbErr).
				Str("original_error", err.Error()).
				Msg("failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Master.Close()
}
