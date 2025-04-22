package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Registration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", h.Registration)

	tests := []struct {
		name         string
		input        dto.RegistrationUserReq
		mockBehavior func()
		expectedCode int
		expectedBody any
	}{
		{
			name: "Success",
			input: dto.RegistrationUserReq{
				Name:        "Test User",
				Email:       "test@example.com",
				Password:    "password123",
				Phone:       "1234567890",
				Fingerprint: "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), "fp123").
					Return("access_token", "refresh_token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: dto.RegistrationUserResp{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
		},
		{
			name: "Invalid Request - Missing Required Fields",
			input: dto.RegistrationUserReq{
				Email:       "test@example.com",
				Password:    "password123",
				Fingerprint: "fp123",
			},
			mockBehavior: func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			input: dto.RegistrationUserReq{
				Name:        "Test User",
				Email:       "test@example.com",
				Password:    "password123",
				Phone:       "1234567890",
				Fingerprint: "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), "fp123").
					Return("", "", errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonData, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != nil {
				var got dto.RegistrationUserResp
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, got)
			}
		})
	}
}

func TestUserHandler_Registration_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", h.Registration)

	jsonData := []byte(`{"invalid json"`)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", h.Login)

	tests := []struct {
		name         string
		input        dto.LoginUserReq
		mockBehavior func()
		expectedCode int
		expectedBody any
	}{
		{
			name: "Success",
			input: dto.LoginUserReq{
				Email:       "test@example.com",
				Password:    "password123",
				Fingerprint: "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "password123", "fp123").
					Return("access_token", "refresh_token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: dto.LoginUserResp{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
		},
		{
			name: "Authentication Error",
			input: dto.LoginUserReq{
				Email:       "test@example.com",
				Password:    "wrong_password",
				Fingerprint: "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "wrong_password", "fp123").
					Return("", "", apperror.ErrIncorrectPassword)
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonData, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != nil {
				var got dto.LoginUserResp
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, got)
			}
		})
	}
}

func TestUserHandler_Login_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", h.Login)

	jsonData := []byte(`{"invalid json"`)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", h.Refresh)

	tests := []struct {
		name         string
		input        dto.RefreshReq
		setCookie    bool
		cookieValue  string
		mockBehavior func()
		expectedCode int
		expectedBody any
	}{
		{
			name: "Success",
			input: dto.RefreshReq{
				RefreshToken: "old_refresh_token",
				Fingerprint:  "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					Refresh(gomock.Any(), "old_refresh_token", "fp123").
					Return("new_access_token", "new_refresh_token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: dto.RefreshResp{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
		},
		{
			name: "Empty Token With Valid Cookie",
			input: dto.RefreshReq{
				RefreshToken: "",
				Fingerprint:  "fp123",
			},
			setCookie:   true,
			cookieValue: "cookie_refresh_token",
			mockBehavior: func() {
				mockService.EXPECT().
					Refresh(gomock.Any(), "cookie_refresh_token", "fp123").
					Return("new_access_token", "new_refresh_token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: dto.RefreshResp{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
		},
		{
			name: "Empty Token Without Cookie",
			input: dto.RefreshReq{
				RefreshToken: "",
				Fingerprint:  "fp123",
			},
			setCookie:    false,
			mockBehavior: func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			input: dto.RefreshReq{
				RefreshToken: "invalid_token",
				Fingerprint:  "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					Refresh(gomock.Any(), "invalid_token", "fp123").
					Return("", "", errors.New("invalid refresh token"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookieValue,
				})
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != nil {
				var got dto.RefreshResp
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, got)
			}
		})
	}
}

func TestUserHandler_Refresh_EmptyToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", h.Refresh)

	tests := []struct {
		name         string
		input        dto.RefreshReq
		setCookie    bool
		cookieValue  string
		mockBehavior func()
		expectedCode int
		expectedBody any
	}{
		{
			name: "Empty Token With Valid Cookie",
			input: dto.RefreshReq{
				RefreshToken: "",
				Fingerprint:  "fp123",
			},
			setCookie:   true,
			cookieValue: "cookie_refresh_token",
			mockBehavior: func() {
				mockService.EXPECT().
					Refresh(gomock.Any(), "cookie_refresh_token", "fp123").
					Return("new_access_token", "new_refresh_token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: dto.RefreshResp{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
		},
		{
			name: "Empty Token Without Cookie",
			input: dto.RefreshReq{
				RefreshToken: "",
				Fingerprint:  "fp123",
			},
			setCookie:    false,
			mockBehavior: func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookieValue,
				})
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != nil {
				var got dto.RefreshResp
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, got)
			}
		})
	}
}

func TestUserHandler_Refresh_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", h.Refresh)

	jsonData := []byte(`{"invalid json"`)
	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", h.Logout)

	tests := []struct {
		name         string
		input        dto.LogoutReq
		mockBehavior func()
		expectedCode int
	}{
		{
			name: "Success",
			input: dto.LogoutReq{
				RefreshToken: "refresh_token",
				Fingerprint:  "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					Logout(gomock.Any(), "refresh_token", "fp123").
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Logout Error",
			input: dto.LogoutReq{
				RefreshToken: "invalid_token",
				Fingerprint:  "fp123",
			},
			mockBehavior: func() {
				mockService.EXPECT().
					Logout(gomock.Any(), "invalid_token", "fp123").
					Return(errors.New("logout failed"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonData, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_Logout_EmptyToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", h.Logout)

	tests := []struct {
		name         string
		input        dto.LogoutReq
		setCookie    bool
		cookieValue  string
		mockBehavior func()
		expectedCode int
	}{
		{
			name: "Empty Token With Valid Cookie",
			input: dto.LogoutReq{
				RefreshToken: "",
				Fingerprint:  "fp123",
			},
			setCookie:   true,
			cookieValue: "cookie_refresh_token",
			mockBehavior: func() {
				mockService.EXPECT().
					Logout(gomock.Any(), "cookie_refresh_token", "fp123").
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Empty Token Without Cookie",
			input: dto.LogoutReq{
				RefreshToken: "",
				Fingerprint:  "fp123",
			},
			setCookie:    false,
			mockBehavior: func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookieValue,
				})
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestUserHandler_Logout_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockUserService(ctrl)
	h := NewUserHandler(mockService, time.Hour, "/api", "example.com")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", h.Logout)

	jsonData := []byte(`{"invalid json"`)
	req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
