package models

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in progress"
	StatusFailed     Status = "failed"
	StatusCompleted  Status = "completed"
)

type Task struct {
	ID          uint
	UserID      uint
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Deadline    time.Time
	Status      Status
}

func NewTask(name, description string, deadline time.Time) (*Task, error) {
	if len(name) > 30 {
		return nil, fmt.Errorf("name should be shorter than 30 characters")
	}

	if len(description) > 150 {
		return nil, fmt.Errorf("description should be shorter than 150 characters")
	}

	if deadline.Before(time.Now()) {
		return nil, fmt.Errorf("deadline can't be earlier than current time")
	}

	task := &Task{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		Deadline:    deadline,
		Status:      StatusPending,
	}

	return task, nil
}
