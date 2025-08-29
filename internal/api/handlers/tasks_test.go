package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"to-do-list/internal/api/types"
	"to-do-list/internal/middleware"
	"to-do-list/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	task.ID = 1
	return args.Error(0)
}

func (m *MockTaskRepository) GetTaskByID(ctx context.Context, id uint) (models.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTasksByUserID(ctx context.Context, userID uint, limit, offset int) ([]models.Task, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Task), args.Error(1)
}
func (m *MockTaskRepository) UpdateTaskName(ctx context.Context, id uint, name string) error {
	args := m.Called(ctx, id, name)
	return args.Error(0)
}
func (m *MockTaskRepository) UpdateTaskDescription(ctx context.Context, id uint, description string) error {
	args := m.Called(ctx, id, description)
	return args.Error(0)
}
func (m *MockTaskRepository) UpdateTaskStatus(ctx context.Context, id uint, status models.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
func (m *MockTaskRepository) DeleteTask(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func withUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, userID)
}

func TestTaskHandler_CreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	handler := NewTaskHandler(mockRepo)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("CreateTask", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Once()

		createReq := types.CreateTaskRequest{
			Name:        "Test Task",
			Description: "Test Description",
			Deadline:    time.Now().Add(24 * time.Hour),
		}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
		ctx := withUserID(req.Context(), 1)
		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp models.Task
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, createReq.Name, resp.Name)
		assert.Equal(t, uint(1), resp.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		createReq := types.CreateTaskRequest{
			Description: "some description",
		}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
		ctx := withUserID(req.Context(), 1)
		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockRepo.AssertNotCalled(t, "CreateTask")
	})
}

func TestTaskHandler_GetTaskByID(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	handler := NewTaskHandler(mockRepo)

	t.Run("Success", func(t *testing.T) {
		taskID := uint(1)
		userID := uint(10)
		mockTask := models.Task{ID: taskID, UserID: userID, Name: "My Task"}
		mockRepo.On("GetTaskByID", mock.Anything, taskID).Return(mockTask, nil).Once()

		req := httptest.NewRequest("GET", "/tasks/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("taskID", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		ctx = withUserID(ctx, userID)

		rr := httptest.NewRecorder()
		handler.GetTaskByID(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp models.Task
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, mockTask.Name, resp.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Forbidden - Another User Task", func(t *testing.T) {
		taskID := uint(1)
		ownerUserID := uint(10)
		requesterUserID := uint(20)
		mockTask := models.Task{ID: taskID, UserID: ownerUserID, Name: "Not Your Task"}
		mockRepo.On("GetTaskByID", mock.Anything, taskID).Return(mockTask, nil).Once()

		req := httptest.NewRequest("GET", "/tasks/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("taskID", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		ctx = withUserID(ctx, requesterUserID)

		rr := httptest.NewRecorder()
		handler.GetTaskByID(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		taskID := uint(999)
		userID := uint(10)
		mockRepo.On("GetTaskByID", mock.Anything, taskID).Return(models.Task{}, errors.New("not found")).Once()

		req := httptest.NewRequest("GET", "/tasks/999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("taskID", "999")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		ctx = withUserID(ctx, userID)

		rr := httptest.NewRecorder()
		handler.GetTaskByID(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}
