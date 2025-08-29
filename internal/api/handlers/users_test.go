package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"to-do-list/internal/api/types"
	"to-do-list/internal/auth"
	"to-do-list/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	user.ID = 1
	user.CreatedAt = time.Now()
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

func TestUserHandler_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	handler := NewUserHandler(mockRepo, tokenManager)

	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	registerReq := types.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp types.AuthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, registerReq.Username, resp.User.Username)
	assert.Equal(t, registerReq.Email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, uint(1), resp.User.ID)

	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Register_ValidationFailure(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	handler := NewUserHandler(mockRepo, tokenManager)

	registerReq := types.RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Contains(t, errResp["error"], "Validation failed")

	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestUserHandler_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	handler := NewUserHandler(mockRepo, tokenManager)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockUser := models.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}
	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

	loginReq := types.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp types.AuthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, mockUser.Username, resp.User.Username)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	handler := NewUserHandler(mockRepo, tokenManager)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockUser := models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

	loginReq := types.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute*15)
	handler := NewUserHandler(mockRepo, tokenManager)

	mockRepo.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(models.User{}, assert.AnError)

	loginReq := types.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockRepo.AssertExpectations(t)
}
