package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/wb-go/wbf/ginext"

	"EventBooker/internal/model"
	"EventBooker/internal/service"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingService(s service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: s}
}

func (h *BookingHandler) Book(c *ginext.Context) {
	userID := c.GetInt("userID")

	eventIDStr := c.Param("event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	b := model.BookingInCreate{
		UserID:  userID,
		EventID: eventID,
	}

	err = h.bookingService.Book(context.Background(), b)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	NewSuccessResponse(c, http.StatusCreated, "book")

}

func (h *BookingHandler) Confirm(c *ginext.Context) {
	userID := c.GetInt("userID")
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.bookingService.Confirm(context.Background(), eventID, userID)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	NewSuccessResponse(c, http.StatusCreated, "book confirmed")

}

func (h *BookingHandler) GetListBooking(c *ginext.Context) {
	userID := c.GetInt("userID")
	lastCreatedAtStr := c.Query("last_created_at")
	lastCreatedAt, err := time.Parse(time.RFC3339, lastCreatedAtStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println(lastCreatedAt)

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

	req := model.BookingGetRequest{
		UserID:        userID,
		Mode:          mode,
		LastCreatedAt: lastCreatedAt,
		LastID:        lastID,
		PageSize:      pageSize,
	}

	b, err := h.bookingService.GetByUserID(context.Background(), req)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	count, err := h.bookingService.GetCountUserBooking(context.Background(), userID)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ginext.H{
		"count":   count,
		"booking": b,
	})
}

func (h *BookingHandler) Cancel(c *ginext.Context) {
	userID := c.GetInt("userID")
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.bookingService.CancelBook(context.Background(), eventID, userID)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	NewSuccessResponse(c, http.StatusOK, "booking canceled")
}
