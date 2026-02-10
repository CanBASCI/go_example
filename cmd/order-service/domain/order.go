package domain

import (
	"time"

	"github.com/google/uuid"

	"go_example/internal/events"
)

// Order represents an order entity.
type Order struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Amount    int64
	Status    events.OrderStatus
	CreatedAt time.Time
}
