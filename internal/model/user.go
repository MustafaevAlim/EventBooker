package model

import "time"

type UserInCreate struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TgChatID *int   `json:"tg_chatid,omitempty"`
}

type UserInRepo struct {
	ID        int
	Email     string
	Password  string
	Role      string
	TgChatID  *int
	CreatedAt time.Time
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserInResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	TgChatID  *int      `json:"tg_chatid"`
	CreatedAt time.Time `json:"created_at"`
}

type UserGetRequest struct {
	LastCreatedAt time.Time
	LastID        int
	Mode          string
	PageSize      int
}
