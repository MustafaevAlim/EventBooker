package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wb-go/wbf/ginext"

	"EventBooker/internal/model"
	"EventBooker/internal/service/mocks"
)

func setupEventTestRouter() *ginext.Engine {
	router := ginext.New("release")
	return router
}

func TestCreateEvent(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    model.EventInCreate
		setupMocks     func(ms *mocks.MockEventService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success create event",
			requestBody: model.EventInCreate{
				Title:              "Test Conference",
				Description:        "Annual developer conference",
				EventDate:          time.Now().Add(48 * time.Hour),
				TotalPlace:         100,
				ReservationPeriod:  "24h",
				BookingConfimation: true,
			},
			setupMocks: func(ms *mocks.MockEventService) {
				ms.On("CreateEvent", mock.Anything, mock.MatchedBy(func(e model.EventInCreate) bool {
					return e.Title == "Test Conference" &&
						e.TotalPlace == 100 &&
						e.BookingConfimation
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"event created"`,
		},

		{
			name: "service error",
			requestBody: model.EventInCreate{
				Title:              "Error Event",
				Description:        "This should fail",
				EventDate:          time.Now().Add(24 * time.Hour),
				TotalPlace:         50,
				ReservationPeriod:  "12h",
				BookingConfimation: false,
			},
			setupMocks: func(ms *mocks.MockEventService) {
				ms.On("CreateEvent", mock.Anything, mock.Anything).Return(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockEventService(t)
			handler := NewEventHandler(mockService)
			router := setupEventTestRouter()

			router.POST("/events", handler.CreateEvent)

			tt.setupMocks(mockService)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected text")

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetEvent(t *testing.T) {
	tests := []struct {
		name           string
		eventIDStr     string
		setupMocks     func(ms *mocks.MockEventService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "success get event",
			eventIDStr: "15",
			setupMocks: func(ms *mocks.MockEventService) {
				expectedEvent := model.EventInResponse{
					ID:                 15,
					Title:              "Test Event",
					Description:        "Event description",
					EventDate:          time.Now().Add(24 * time.Hour),
					TotalPlace:         200,
					OccupiedPlace:      50,
					EventStatus:        model.EventStatusPending,
					ReservationPeriod:  "24h",
					BookingConfimation: true,
					CreatedAt:          time.Now(),
				}
				ms.On("GetByID", mock.Anything, 15).Return(expectedEvent, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"title":"Test Event"`,
		},
		{
			name:           "invalid event id",
			eventIDStr:     "abc",
			setupMocks:     func(ms *mocks.MockEventService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid syntax",
		},
		{
			name:       "service error - event not found",
			eventIDStr: "999",
			setupMocks: func(ms *mocks.MockEventService) {
				ms.On("GetByID", mock.Anything, 999).Return(model.EventInResponse{}, errors.New("event not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "event not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockEventService(t)
			handler := NewEventHandler(mockService)
			router := setupEventTestRouter()

			router.GET("/events/:id", handler.GetEvent)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/events/"+tt.eventIDStr, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected text")

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetListEvents_QueryParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMocks     func(ms *mocks.MockEventService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success get events list",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "10",
				"mode":            "all",
			},
			setupMocks: func(ms *mocks.MockEventService) {
				events := []model.EventInResponse{
					{
						ID:                 1,
						Title:              "First Event",
						Description:        "First event description",
						EventDate:          time.Now().Add(48 * time.Hour),
						TotalPlace:         150,
						OccupiedPlace:      30,
						EventStatus:        model.EventStatusPending,
						ReservationPeriod:  "24h",
						BookingConfimation: true,
						CreatedAt:          time.Now(),
					},
				}
				ms.On("GetListEvents", mock.Anything, mock.MatchedBy(func(req model.EventGetRequest) bool {
					return req.LastID == 0 &&
						req.PageSize == 10 &&
						req.Mode == "all" &&
						!req.LastCreatedAt.IsZero()
				})).Return(events, nil)

				ms.On("GetCountEvent", mock.Anything).Return(25, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"count":25`,
		},
		{
			name: "invalid time format",
			queryParams: map[string]string{
				"last_created_at": "invalid-time",
				"last_id":         "0",
				"page_size":       "10",
				"mode":            "all",
			},
			setupMocks:     func(ms *mocks.MockEventService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "parsing time",
		},
		{
			name: "invalid last_id",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "abc",
				"page_size":       "10",
				"mode":            "all",
			},
			setupMocks:     func(ms *mocks.MockEventService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "strconv.Atoi",
		},
		{
			name: "invalid page_size",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "xyz",
				"mode":            "all",
			},
			setupMocks:     func(ms *mocks.MockEventService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "strconv.Atoi",
		},
		{
			name: "service error - get list",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "10",
				"mode":            "all",
			},
			setupMocks: func(ms *mocks.MockEventService) {
				ms.On("GetListEvents", mock.Anything, mock.Anything).Return([]model.EventInResponse{}, errors.New("database query failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database query failed",
		},
		{
			name: "service error - get count",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "10",
				"mode":            "all",
			},
			setupMocks: func(ms *mocks.MockEventService) {
				events := []model.EventInResponse{{ID: 1, Title: "Test"}}
				ms.On("GetListEvents", mock.Anything, mock.Anything).Return(events, nil)
				ms.On("GetCountEvent", mock.Anything).Return(0, errors.New("count query failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "count query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockEventService(t)
			handler := NewEventHandler(mockService)
			router := setupEventTestRouter()

			router.GET("/events", handler.GetListEvents)

			tt.setupMocks(mockService)

			testURL := buildURL("/events", tt.queryParams)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", testURL, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected text")
			}

			mockService.AssertExpectations(t)
		})
	}
}
