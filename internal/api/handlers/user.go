package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/wb-go/wbf/ginext"

	"EventBooker/internal/model"
	"EventBooker/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{userService: s}
}

func (h *UserHandler) Register(c *ginext.Context) {
	var u model.UserInCreate

	err := c.ShouldBindJSON(&u)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.userService.CreateUser(context.Background(), u)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, ginext.H{
		"token": token,
	})

}

func (h *UserHandler) Login(c *ginext.Context) {
	var u model.UserLoginRequest

	err := c.ShouldBindJSON(&u)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.userService.Login(context.Background(), u)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ginext.H{
		"token": token,
	})
}

func (h *UserHandler) GetList(c *ginext.Context) {
	lastCreatedAtStr := c.Query("last_created_at")
	lastCreatedAt, err := time.Parse(time.RFC3339, lastCreatedAtStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	lastIDStr := c.Query("last_id")
	lastID, err := strconv.Atoi(lastIDStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	pageSizeStr := c.Query("page_size")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	mode := c.Query("mode")

	req := model.UserGetRequest{
		Mode:          mode,
		LastCreatedAt: lastCreatedAt,
		LastID:        lastID,
		PageSize:      pageSize,
	}

	u, err := h.userService.GetListUsers(context.Background(), req)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	count, err := h.userService.GetCountUsers(context.Background())
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ginext.H{
		"count": count,
		"users": u,
	})
}
