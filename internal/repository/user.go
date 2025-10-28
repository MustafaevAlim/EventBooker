package repository

import (
	"context"
	"time"

	"github.com/wb-go/wbf/zlog"

	"EventBooker/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u model.UserInCreate) error
	GetByID(ctx context.Context, id int) (model.UserInRepo, error)
	GetByEmail(ctx context.Context, email string) (model.UserInRepo, error)
	GetListUsers(ctx context.Context, req model.UserGetRequest) ([]model.UserInRepo, error)
	GetCountUsers(ctx context.Context) (int, error)
}

type userRepository struct {
	db dbInterface
}

func NewUserRepository(db dbInterface) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) Create(ctx context.Context, u model.UserInCreate) error {
	query := `INSERT INTO users (email, password, role, tg_chatid, created_at)
				VALUES ($1, $2, $3, $4, $5)`

	_, err := ur.db.ExecContext(ctx, query, u.Email, u.Password, "user", u.TgChatID, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (ur *userRepository) GetByID(ctx context.Context, id int) (model.UserInRepo, error) {
	query := `SELECT *
				FROM users
				WHERE user_id = $1`
	res, err := ur.db.QueryContext(ctx, query, id)
	if err != nil {
		return model.UserInRepo{}, err
	}

	var record model.UserInRepo
	if res.Next() {
		err := res.Scan(&record.ID, &record.Email, &record.Password, &record.Role, &record.TgChatID, &record.CreatedAt)
		if err != nil {
			return model.UserInRepo{}, err
		}
	}
	return record, nil
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (model.UserInRepo, error) {
	query := `SELECT *
				FROM users
				WHERE email = $1`
	res, err := ur.db.QueryContext(ctx, query, email)
	if err != nil {
		return model.UserInRepo{}, err
	}

	var record model.UserInRepo
	if res.Next() {
		err := res.Scan(&record.ID, &record.Email, &record.Password, &record.Role, &record.TgChatID, &record.CreatedAt)
		if err != nil {
			return model.UserInRepo{}, err
		}
	}
	return record, nil
}

func (ur *userRepository) GetListUsers(ctx context.Context, req model.UserGetRequest) ([]model.UserInRepo, error) {
	var query string
	switch req.Mode {
	case "next":
		query = `SELECT *
					FROM users
					WHERE created_at > $1 AND user_id > $2
					ORDER BY created_at ASC, user_id ASC
					LIMIT $3`
	case "prev":
		query = `SELECT *
					FROM users
					WHERE (created_at < $1) OR (created_at = $1 AND user_id < $2)
					ORDER BY created_at DESC, user_id DESC
					LIMIT $3`
	}
	args := []any{req.LastCreatedAt, req.LastID, req.PageSize}

	res, err := ur.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			zlog.Logger.Error().Msg(err.Error())
		}
	}()

	var e []model.UserInRepo
	for res.Next() {
		var temp model.UserInRepo
		err := res.Scan(&temp.ID, &temp.Email, &temp.Password, &temp.Role, &temp.TgChatID, &temp.CreatedAt)
		if err != nil {
			return nil, err
		}
		e = append(e, temp)
	}
	return e, nil
}

func (ur *userRepository) GetCountUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*)
				FROM users`
	res := ur.db.QueryRowContext(ctx, query)
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
