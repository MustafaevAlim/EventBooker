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

func setupUserTestRouter() *ginext.Engine {
	return ginext.New("release")
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    model.UserInCreate
		setupMocks     func(ms *mocks.MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success register",
			requestBody: model.UserInCreate{
				Email:    "test@mail.com",
				Password: "securepass",
			},
			setupMocks: func(ms *mocks.MockUserService) {
				ms.On("CreateUser", mock.Anything, mock.MatchedBy(func(u model.UserInCreate) bool {
					return u.Email == "test@mail.com"
				})).Return("jwt_token_test", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"token":"jwt_token_test"`,
		},

		{
			name: "service error",
			requestBody: model.UserInCreate{
				Email:    "err@mail.com",
				Password: "pass",
			},
			setupMocks: func(ms *mocks.MockUserService) {
				ms.On("CreateUser", mock.Anything, mock.Anything).Return("", errors.New("db failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "db failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			handler := NewUserHandler(mockService)
			router := setupUserTestRouter()

			router.POST("/register", handler.Register)
			tt.setupMocks(mockService)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    model.UserLoginRequest
		setupMocks     func(ms *mocks.MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success login",
			requestBody: model.UserLoginRequest{
				Email:    "user@mail.com",
				Password: "mypassword",
			},
			setupMocks: func(ms *mocks.MockUserService) {
				ms.On("Login", mock.Anything, mock.MatchedBy(func(u model.UserLoginRequest) bool {
					return u.Email == "user@mail.com"
				})).Return("jwt_token_login", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"token":"jwt_token_login"`,
		},

		{
			name: "login error",
			requestBody: model.UserLoginRequest{
				Email:    "unknown@mail.com",
				Password: "nopass",
			},
			setupMocks: func(ms *mocks.MockUserService) {
				ms.On("Login", mock.Anything, mock.Anything).Return("", errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			handler := NewUserHandler(mockService)
			router := setupUserTestRouter()

			router.POST("/login", handler.Login)
			tt.setupMocks(mockService)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetList(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMocks     func(ms *mocks.MockUserService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			queryParams: map[string]string{
				"last_created_at": now.Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "5",
				"mode":            "all",
			},
			setupMocks: func(ms *mocks.MockUserService) {
				users := []model.UserInResponse{
					{ID: 1, Email: "a@mail.com", TgChatID: nil, CreatedAt: now},
					{ID: 2, Email: "b@mail.com", TgChatID: nil, CreatedAt: now.Add(-time.Hour)},
				}
				ms.On("GetListUsers", mock.Anything, mock.MatchedBy(func(req model.UserGetRequest) bool {
					return req.PageSize == 5 && req.Mode == "all"
				})).Return(users, nil)
				ms.On("GetCountUsers", mock.Anything).Return(2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"count":2`,
		},
		{
			name: "invalid time",
			queryParams: map[string]string{
				"last_created_at": "badtime",
				"last_id":         "0",
				"page_size":       "5",
				"mode":            "all",
			},
			setupMocks:     func(ms *mocks.MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "parsing time",
		},
		{
			name: "service error list",
			queryParams: map[string]string{
				"last_created_at": now.Format(time.RFC3339),
				"last_id":         "0",
				"page_size":       "5",
				"mode":            "all",
			},
			setupMocks: func(ms *mocks.MockUserService) {
				ms.On("GetListUsers", mock.Anything, mock.Anything).Return([]model.UserInResponse{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			handler := NewUserHandler(mockService)
			router := setupUserTestRouter()

			router.GET("/users", handler.GetList)
			tt.setupMocks(mockService)

			testURL := buildURL("/users", tt.queryParams)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", testURL, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}
