package types

import (
	"time"
	"to-do-list/internal/models"
)

type CreateTaskRequest struct {
	Name        string    `json:"name" validate:"required,max=30"`
	Description string    `json:"description" validate:"max=150"`
	Deadline    time.Time `json:"deadline" validate:"required"`
}

type UpdateTaskRequest struct {
	Name        *string        `json:"name" validate:"omitempty,max=30"`
	Description *string        `json:"description" validate:"omitempty,max=150"`
	Deadline    *time.Time     `json:"deadline" validate:"omitempty"`
	Status      *models.Status `json:"status" validate:"omitempty,oneof=pending 'in progress' failed completed"`
}

type TaskResponse struct {
	ID          uint          `json:"id"`
	UserID      uint          `json:"userId"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"createdAt"`
	Deadline    time.Time     `json:"deadline"`
	Status      models.Status `json:"status"`
}
