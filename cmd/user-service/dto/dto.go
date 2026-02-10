package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Username       string `json:"username"`
	InitialBalance int64  `json:"initialBalance"`
}

// UserResponse is the user API response.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}
