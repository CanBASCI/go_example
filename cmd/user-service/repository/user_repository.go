package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go_example/cmd/user-service/domain"
)

// UserRepository handles user persistence.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, username, balance, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, u.ID, u.Username, u.Balance, u.CreatedAt)
	return err
}

// GetByID returns a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, balance, created_at FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Username, &u.Balance, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateBalance updates user balance (for reserve/release credit).
func (r *UserRepository) UpdateBalance(ctx context.Context, id uuid.UUID, balance int64) error {
	query := `UPDATE users SET balance = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, balance, id)
	return err
}
