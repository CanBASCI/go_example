// Package events defines shared Kafka event DTOs for the saga choreography.
package events

import "github.com/google/uuid"

// OrderStatus represents order status in the saga.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusCanceled  OrderStatus = "CANCELED"
)

// OrderCreatedEvent is published when an order is created.
// User-service listens to reserve credit.
type OrderCreatedEvent struct {
	OrderID uuid.UUID `json:"orderId"`
	UserID  uuid.UUID `json:"userId"`
	Amount  int64     `json:"amount"`
}

// OrderCanceledEvent is published when an order is canceled.
// User-service listens to release reserved credit (compensation).
type OrderCanceledEvent struct {
	OrderID uuid.UUID `json:"orderId"`
	UserID  uuid.UUID `json:"userId"`
	Amount  int64     `json:"amount"`
}

// UserCreditReservedEvent is published when credit is successfully reserved.
// Order-service listens to confirm the order.
type UserCreditReservedEvent struct {
	OrderID uuid.UUID `json:"orderId"`
	UserID  uuid.UUID `json:"userId"`
	Amount  int64     `json:"amount"`
}

// UserCreditReservationFailedEvent is published when credit reservation fails.
// Order-service listens to cancel the order.
type UserCreditReservationFailedEvent struct {
	OrderID uuid.UUID `json:"orderId"`
	UserID  uuid.UUID `json:"userId"`
	Amount  int64     `json:"amount"`
	Reason  string    `json:"reason"`
}
