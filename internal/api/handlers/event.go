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

type EventHandler struct {
	eventService service.EventService
}

func NewEventHandler(s service.EventService) *EventHandler {
	return &EventHandler{eventService: s}
}

func (h *EventHandler) CreateEvent(c *ginext.Context) {
	var e model.EventInCreate

	err := c.ShouldBindJSON(&e)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.eventService.CreateEvent(context.Background(), e)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	NewSuccessResponse(c, http.StatusCreated, "event created")
}

func (h *EventHandler) GetEvent(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	e, err := h.eventService.GetByID(context.Background(), id)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, e)
}

func (h *EventHandler) GetListEvents(c *ginext.Context) {
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

	req := model.EventGetRequest{
		LastCreatedAt: lastCreatedAt,
		LastID:        lastID,
		Mode:          mode,
		PageSize:      pageSize,
	}

	e, err := h.eventService.GetListEvents(context.Background(), req)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	countEvent, err := h.eventService.GetCountEvent(context.Background())
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ginext.H{
		"count":  countEvent,
		"events": e,
	})
}
