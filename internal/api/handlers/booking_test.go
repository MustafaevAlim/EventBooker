package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wb-go/wbf/ginext"

	"EventBooker/internal/model"
	"EventBooker/internal/service/mocks"
)

func setupTestRouter(userID int) *ginext.Engine {
	router := ginext.New("release")
	router.Use(func(c *ginext.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	return router
}

func buildURL(basePath string, params map[string]string) string {
	u := &url.URL{Path: basePath}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func TestBookHandler_Book(t *testing.T) {
	tests := []struct {
		name           string
		eventIDStr     string
		userID         int
		setupMocks     func(ms *mocks.MockBookingService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "success booking",
			eventIDStr: "15",
			userID:     42,
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("Book", mock.Anything, mock.MatchedBy(func(b model.BookingInCreate) bool {
					return b.UserID == 42 && b.EventID == 15
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"book"`,
		},
		{
			name:           "invalid event id",
			eventIDStr:     "invalid",
			userID:         42,
			setupMocks:     func(ms *mocks.MockBookingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid syntax",
		},
		{
			name:       "service error",
			eventIDStr: "10",
			userID:     42,
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("Book", mock.Anything, mock.Anything).Return(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database connection failed",
		},
		{
			name:       "zero user id",
			eventIDStr: "5",
			userID:     0,
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("Book", mock.Anything, mock.MatchedBy(func(b model.BookingInCreate) bool {
					return b.UserID == 0 && b.EventID == 5
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"book"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockBookingService(t)
			handler := NewBookingService(mockService)
			router := setupTestRouter(tt.userID)
			router.POST("/book/:event_id", handler.Book)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/book/"+tt.eventIDStr, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected text")

			mockService.AssertExpectations(t)
		})
	}
}

func TestBookHandler_Confirm(t *testing.T) {
	tests := []struct {
		name           string
		eventIDStr     string
		userID         int
		setupMocks     func(ms *mocks.MockBookingService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "success confirm",
			eventIDStr: "25",
			userID:     42,
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("Confirm", mock.Anything, 25, 42).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"book confirmed"`,
		},
		{
			name:           "invalid event id",
			eventIDStr:     "abc",
			userID:         42,
			setupMocks:     func(ms *mocks.MockBookingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid syntax",
		},

		{
			name:       "service error",
			eventIDStr: "7",
			userID:     42,
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("Confirm", mock.Anything, 7, 42).Return(errors.New("event not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "event not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockBookingService(t)
			handler := NewBookingService(mockService)
			router := setupTestRouter(tt.userID)
			router.POST("/confirm/:event_id", handler.Confirm)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/confirm/"+tt.eventIDStr, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected text")

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetListBookingHandler_QueryParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		userID         int
		setupMocks     func(ms *mocks.MockBookingService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "invalid time format",
			queryParams: map[string]string{
				"last_created_at": "invalid-time",
				"last_id":         "0",
				"page_size":       "10",
				"mode":            "my",
			},
			userID:         42,
			setupMocks:     func(ms *mocks.MockBookingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "parsing time",
		},
		{
			name: "invalid last_id",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "invalid",
				"page_size":       "10",
				"mode":            "my",
			},
			userID:         42,
			setupMocks:     func(ms *mocks.MockBookingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "strconv.Atoi",
		},
		{
			name: "invalid page_size",
			queryParams: map[string]string{
				"last_created_at": time.Now().Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "invalid",
				"mode":            "my",
			},
			userID:         42,
			setupMocks:     func(ms *mocks.MockBookingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "strconv.Atoi",
		},
		{
			name: "missing parameters - should use defaults",
			queryParams: map[string]string{
				"mode": "my",
			},
			userID: 42,
			setupMocks: func(ms *mocks.MockBookingService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "parsing time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockBookingService(t)
			handler := NewBookingService(mockService)
			router := setupTestRouter(tt.userID)

			queryParams := tt.queryParams
			testURL := buildURL("/bookings", queryParams)

			router.GET("/bookings", handler.GetListBooking)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", testURL, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected error")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetListBookingHandler_Success(t *testing.T) {
	mockService := mocks.NewMockBookingService(t)
	handler := NewBookingService(mockService)
	router := setupTestRouter(42)

	now := time.Now().UTC()

	bookings := []model.BookingInResponse{
		{
			ID:               1,
			UserID:           42,
			EventID:          10,
			Status:           model.StatusBookingConfirmed,
			ExpiresAt:        now.Add(24 * time.Hour),
			CreatedAt:        now,
			EventTitle:       "Test Conference",
			EventDescription: "Annual developer conference",
			EventDate:        now.Add(48 * time.Hour),
		},
		{
			ID:               2,
			UserID:           42,
			EventID:          11,
			Status:           model.StatusBookingPending,
			ExpiresAt:        now.Add(2 * time.Hour),
			CreatedAt:        now.Add(-1 * time.Hour),
			EventTitle:       "Workshop",
			EventDescription: "Hands-on workshop",
			EventDate:        now.Add(72 * time.Hour),
		},
	}

	mockService.On("GetByUserID", mock.Anything, mock.MatchedBy(func(req model.BookingGetRequest) bool {
		return req.UserID == 42 &&
			req.LastID == 0 &&
			req.PageSize == 10 &&
			req.Mode == "my" &&
			!req.LastCreatedAt.IsZero()
	})).Return(bookings, nil)

	mockService.On("GetCountUserBooking", mock.Anything, 42).Return(2, nil)

	router.GET("/bookings", handler.GetListBooking)

	queryParams := map[string]string{
		"last_created_at": now.Format(time.RFC3339),
		"last_id":         "0",
		"page_size":       "10",
		"mode":            "my",
	}
	testURL := buildURL("/bookings", queryParams)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", testURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Count   int                       `json:"count"`
		Booking []model.BookingInResponse `json:"booking"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Booking, 2)
	assert.Equal(t, "Test Conference", response.Booking[0].EventTitle)
	assert.Equal(t, "Annual developer conference", response.Booking[0].EventDescription)

	mockService.AssertExpectations(t)
}

func TestGetListBookingHandler_ServiceErrors(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(ms *mocks.MockBookingService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "get by user id error",
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("GetByUserID", mock.Anything, mock.Anything).Return([]model.BookingInResponse{}, errors.New("database query failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database query failed",
		},
		{
			name: "get count error",
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("GetByUserID", mock.Anything, mock.Anything).Return([]model.BookingInResponse{}, nil)
				ms.On("GetCountUserBooking", mock.Anything, 42).Return(0, errors.New("count query failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "count query failed",
		},
	}

	now := time.Now().UTC()
	queryParams := map[string]string{
		"last_created_at": now.Format(time.RFC3339),
		"last_id":         "0",
		"page_size":       "10",
		"mode":            "my",
	}
	testURL := buildURL("/bookings", queryParams)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockBookingService(t)
			handler := NewBookingService(mockService)
			router := setupTestRouter(42)

			router.GET("/bookings", handler.GetListBooking)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", testURL, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			mockService.AssertExpectations(t)
		})
	}
}

func TestCancelHandler(t *testing.T) {
	tests := []struct {
		name           string
		eventIDStr     string
		userID         int
		setupMocks     func(ms *mocks.MockBookingService)
		expectedStatus int
		expectedBody   string
		httpMethod     string
	}{
		{
			name:       "success cancel",
			eventIDStr: "30",
			userID:     42,
			httpMethod: "POST",
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("CancelBook", mock.Anything, 30, 42).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"booking canceled"`,
		},
		{
			name:           "invalid event id",
			eventIDStr:     "notanumber",
			userID:         42,
			httpMethod:     "POST",
			setupMocks:     func(ms *mocks.MockBookingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid syntax",
		},

		{
			name:       "service error",
			eventIDStr: "12",
			userID:     42,
			httpMethod: "POST",
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("CancelBook", mock.Anything, 12, 42).Return(errors.New("booking not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "booking not found",
		},
		{
			name:       "zero user id cancel",
			eventIDStr: "8",
			userID:     0,
			httpMethod: "POST",
			setupMocks: func(ms *mocks.MockBookingService) {
				ms.On("CancelBook", mock.Anything, 8, 0).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"booking canceled"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockBookingService(t)
			handler := NewBookingService(mockService)
			router := setupTestRouter(tt.userID)

			router.POST("/cancel/:event_id", handler.Cancel)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.httpMethod, "/cancel/"+tt.eventIDStr, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code should match")
			assert.Contains(t, w.Body.String(), tt.expectedBody, "response body should contain expected text")

			mockService.AssertExpectations(t)
		})
	}
}
