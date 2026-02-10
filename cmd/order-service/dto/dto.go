package dto

import (
	"time"

	"github.com/google/uuid"

	"go_example/internal/events"
)

// CreateOrderRequest is the request body for creating an order.
type CreateOrderRequest struct {
	UserID uuid.UUID `json:"userId"`
	Amount int64     `json:"amount"`
}

// OrderResponse is the order API response.
type OrderResponse struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"userId"`
	Amount    int64           `json:"amount"`
	Status    events.OrderStatus `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
}
