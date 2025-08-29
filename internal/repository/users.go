package repository

import (
	"context"
	"database/sql"
	"fmt"
	"to-do-list/internal/models"

	_ "github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}


func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: failed to create a user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	query := `SELECT id, username, email, password, created_at FROM users WHERE email = $1`

	row := r.db.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("repository: failed to get user by email: %w", err)
	}

	return user, nil
}
