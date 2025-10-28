package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"

	"EventBooker/internal/config"
	"EventBooker/internal/model"
	"EventBooker/internal/repository"
)

var jwtSecret []byte

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type UserService interface {
	CreateUser(ctx context.Context, u model.UserInCreate) (string, error)
	Login(ctx context.Context, req model.UserLoginRequest) (string, error)
	GetListUsers(ctx context.Context, req model.UserGetRequest) ([]model.UserInResponse, error)
	GetCountUsers(ctx context.Context) (int, error)
}

type userService struct {
	storage *repository.Storage
}

func NewUserService(s *repository.Storage) *userService {
	c, _ := config.NewConfig()
	jwtSecret = []byte(c.Server.JwtKey)
	return &userService{storage: s}
}

func (us *userService) CreateUser(ctx context.Context, u model.UserInCreate) (string, error) {

	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.CreateUser error: %v", err)
		return "", err
	}
	u.Password = hashedPassword

	err = us.storage.User.Create(ctx, u)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.CreateUser error: %v", err)
		return "", err
	}

	user, err := us.storage.User.GetByEmail(ctx, u.Email)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.CreateUser error: %v", err)
		return "", err
	}

	token, err := GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.CreateUser error: %v", err)
		return "", err
	}

	return token, nil

}

func (us *userService) Login(ctx context.Context, req model.UserLoginRequest) (string, error) {
	user, err := us.storage.User.GetByEmail(ctx, req.Email)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.Login error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	if !checkPassword(req.Password, user.Password) {
		return "", ErrUnauthorized
	}
	token, err := GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.Login error: %v", err)
		return "", err
	}

	return token, nil

}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return errors.Is(err, nil)
}

func GenerateToken(userID int, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "event-booker",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

func (us *userService) GetListUsers(ctx context.Context, req model.UserGetRequest) ([]model.UserInResponse, error) {
	usersInRepo, err := us.storage.User.GetListUsers(ctx, req)
	if err != nil {
		zlog.Logger.Error().Msgf("service.UserService.GetListUsers error: %v", err)
		return nil, err
	}

	usersInResponse := make([]model.UserInResponse, 0, len(usersInRepo))
	for _, u := range usersInRepo {
		usersInResponse = append(usersInResponse, model.UserInResponse{
			ID:        u.ID,
			Email:     u.Email,
			TgChatID:  u.TgChatID,
			CreatedAt: u.CreatedAt,
		})
	}

	return usersInResponse, nil

}

func (us *userService) GetCountUsers(ctx context.Context) (int, error) {
	return us.storage.User.GetCountUsers(ctx)
}
