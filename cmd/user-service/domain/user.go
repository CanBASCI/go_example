package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity.
type User struct {
	ID        uuid.UUID
	Username  string
	Balance   int64
	CreatedAt time.Time
}
