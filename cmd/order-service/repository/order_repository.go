package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go_example/internal/events"
	"go_example/cmd/order-service/domain"
)

// OrderRepository handles order persistence.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Create inserts a new order.
func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	query := `INSERT INTO orders (id, user_id, amount, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, o.ID, o.UserID, o.Amount, string(o.Status), o.CreatedAt)
	return err
}

// GetByID returns an order by ID.
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `SELECT id, user_id, amount, status, created_at FROM orders WHERE id = $1`
	var o domain.Order
	var status string
	err := r.pool.QueryRow(ctx, query, id).Scan(&o.ID, &o.UserID, &o.Amount, &status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	o.Status = events.OrderStatus(status)
	return &o, nil
}

// UpdateStatus updates order status.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status events.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, string(status), id)
	return err
}

// ListByUserID returns all orders for a user.
func (r *OrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) {
	query := `SELECT id, user_id, amount, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Order
	for rows.Next() {
		var o domain.Order
		var status string
		if err := rows.Scan(&o.ID, &o.UserID, &o.Amount, &status, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Status = events.OrderStatus(status)
		list = append(list, &o)
	}
	return list, rows.Err()
}
