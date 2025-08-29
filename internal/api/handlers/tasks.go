package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"to-do-list/internal/api/types"
	"to-do-list/internal/middleware"
	"to-do-list/internal/models"
	"to-do-list/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func getUserIDFromCtx(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return 0, errors.New("invalid user ID in context")
	}
	return userID, nil
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	var req types.CreateTaskRequest
	err = render.DecodeJSON(r.Body, &req)
	if err != nil {
		slog.Error("Failed to decode request body", slog.Any("error", err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	err = validate.Struct(req)
	if err != nil {
		slog.Error("Validation failed", slog.Any("error", err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Validation failed: " + err.Error()})
		return
	}

	task, err := models.NewTask(req.Name, req.Description, req.Deadline)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": err.Error()})

		return
	}
	task.UserID = userID

	err = h.repo.CreateTask(r.Context(), task)
	if err != nil {
		slog.Error("Failed to create task", slog.Any("error", err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to create task"})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, task)
}

func (h *TaskHandler) GetUserTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 
	offset := 0 

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	tasks, err := h.repo.GetTasksByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		slog.Error("Failed to get user tasks", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to get tasks"})

		return
	}

	render.JSON(w, r, tasks)
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})

		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid task ID"})
		return
	}

	task, err := h.repo.GetTaskByID(r.Context(), uint(taskID))
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Task not found"})
		return
	}

	if task.UserID != userID {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Task not found"})
		return
	}

	render.JSON(w, r, task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})

		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse task ID", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid task ID"})

		return
	}

	task, err := h.repo.GetTaskByID(r.Context(), uint(taskID))
	if err != nil {
		slog.Error("Failed to get task by ID", slog.Any("error", err))

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Task not found"})

		return
	}

	if task.UserID != userID {
		slog.Error("User attempted to update another user's task", slog.Any("taskID", taskID), slog.Any("userID", userID))

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Task not found"})

		return
	}

	var req types.UpdateTaskRequest
	err = render.DecodeJSON(r.Body, &req)
	if err != nil {
		slog.Error("Failed to decode update request body", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})

		return
	}

	err = validate.Struct(req)
	if err != nil {
		slog.Error("Update validation failed", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Validation failed: " + err.Error()})

		return
	}

	if req.Name != nil {
		err = h.repo.UpdateTaskName(r.Context(), uint(taskID), *req.Name)
		if err != nil {
			slog.Error("Failed to update task name", slog.Any("error", err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Failed to update task name"})

			return
		}
	}
	if req.Description != nil {
		err = h.repo.UpdateTaskDescription(r.Context(), uint(taskID), *req.Description)
		if err != nil {
			slog.Error("Failed to update task description", slog.Any("error", err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Failed to update task description"})

			return
		}
	}
	if req.Status != nil {
		err = h.repo.UpdateTaskStatus(r.Context(), uint(taskID), *req.Status)
		if err != nil {
			slog.Error("Failed to update task status", slog.Any("error", err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Failed to update task status"})

			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})

		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		slog.Error("Failed to parse task ID", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid task ID"})

		return
	}

	task, err := h.repo.GetTaskByID(r.Context(), uint(taskID))
	if err != nil {
		slog.Error("Failed to get task by ID", slog.Any("error", err))

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Task not found"})

		return
	}

	if task.UserID != userID {
		slog.Error("User attempted to delete another user's task", slog.Any("taskID", taskID), slog.Any("userID", userID))

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Task not found"})

		return
	}

	err = h.repo.DeleteTask(r.Context(), uint(taskID))
	if err != nil {
		slog.Error("Failed to delete task", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to delete task"})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
