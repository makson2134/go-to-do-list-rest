package repository

import (
	"context"
	"database/sql"
	"fmt"
	"to-do-list/internal/models"

	_ "github.com/lib/pq"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	UpdateTaskName(ctx context.Context, id uint, name string) error
	UpdateTaskDescription(ctx context.Context, id uint, description string) error
	UpdateTaskStatus(ctx context.Context, id uint, status models.Status) error
	DeleteTask(ctx context.Context, id uint) error
	GetTaskByID(ctx context.Context, id uint) (models.Task, error)
	GetTasksByUserID(ctx context.Context, userID uint, limit, offset int) ([]models.Task, error)
}

type PostgresTaskRepository struct {
	db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
	query := `INSERT INTO tasks (user_id, name, description, created_at, deadline, status)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		task.UserID,
		task.Name,
		task.Description,
		task.CreatedAt,
		task.Deadline,
		task.Status,
	).Scan(&task.ID)

	if err != nil {
		return fmt.Errorf("repository: failed to create a task: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) UpdateTaskName(ctx context.Context, id uint, name string) error {
	query := `UPDATE tasks SET name = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, name, id)

	if err != nil {
		return fmt.Errorf("repository: failed to update task name: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) UpdateTaskDescription(ctx context.Context, id uint, description string) error {
	query := `UPDATE tasks SET description = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, description, id)

	if err != nil {
		return fmt.Errorf("repository: failed to update task description: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) UpdateTaskStatus(ctx context.Context, id uint, status models.Status) error {
	query := `UPDATE tasks SET status = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, status, id)

	if err != nil {
		return fmt.Errorf("repository: failed to update task status: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) DeleteTask(ctx context.Context, id uint) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("repository: failed to delete task: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) GetTaskByID(ctx context.Context, id uint) (models.Task, error) {
	var task models.Task

	query := `SELECT id, user_id, name, description, created_at, updated_at, deadline, status FROM tasks WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.CreatedAt, &task.UpdatedAt, &task.Deadline, &task.Status)
	if err != nil {
		return models.Task{}, fmt.Errorf("repository: failed to get task by id: %w", err)
	}

	return task, nil
}

func (r *PostgresTaskRepository) GetTasksByUserID(ctx context.Context, userID uint, limit, offset int) ([]models.Task, error) {
	query := `SELECT id, user_id, name, description, created_at, updated_at, deadline, status FROM tasks WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get tasks by user id: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.CreatedAt, &task.UpdatedAt, &task.Deadline, &task.Status)
		if err != nil {
			return nil, fmt.Errorf("repository: failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
