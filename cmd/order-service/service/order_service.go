package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go_example/internal/events"
	"go_example/cmd/order-service/domain"
	"go_example/cmd/order-service/dto"
	"go_example/cmd/order-service/repository"
)

var ErrOrderNotFound = errors.New("order not found")

// OrderService implements order business logic and saga coordination.
type OrderService struct {
	repo   *repository.OrderRepository
	writer OrderEventWriter
}

// OrderEventWriter publishes order events to Kafka.
type OrderEventWriter interface {
	PublishOrderCreated(ctx context.Context, evt events.OrderCreatedEvent) error
	PublishOrderCanceled(ctx context.Context, evt events.OrderCanceledEvent) error
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo *repository.OrderRepository, writer OrderEventWriter) *OrderService {
	return &OrderService{repo: repo, writer: writer}
}

// CreateOrder creates an order with PENDING status and publishes OrderCreatedEvent.
func (s *OrderService) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	o := &domain.Order{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Amount:    req.Amount,
		Status:    events.OrderStatusPending,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, err
	}
	evt := events.OrderCreatedEvent{OrderID: o.ID, UserID: o.UserID, Amount: o.Amount}
	if err := s.writer.PublishOrderCreated(ctx, evt); err != nil {
		return nil, err
	}
	return toOrderResponse(o), nil
}

// GetByID returns an order by ID.
func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return toOrderResponse(o), nil
}

// ListByUserID returns all orders for a user.
func (s *OrderService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.OrderResponse, error) {
	orders, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.OrderResponse, len(orders))
	for i, o := range orders {
		out[i] = toOrderResponse(o)
	}
	return out, nil
}

// ConfirmOrder sets order status to CONFIRMED.
func (s *OrderService) ConfirmOrder(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, o.ID, events.OrderStatusConfirmed)
}

// CancelOrder sets status to CANCELED and publishes OrderCanceledEvent (compensation).
func (s *OrderService) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}
		return err
	}
	prev := o.Status
	if err := s.repo.UpdateStatus(ctx, o.ID, events.OrderStatusCanceled); err != nil {
		return err
	}
	if prev == events.OrderStatusPending || prev == events.OrderStatusConfirmed {
		evt := events.OrderCanceledEvent{OrderID: o.ID, UserID: o.UserID, Amount: o.Amount}
		if err := s.writer.PublishOrderCanceled(ctx, evt); err != nil {
			log.Printf("[order-service] PublishOrderCanceled failed (order %s): %v", o.ID, err)
			return err
		}
	}
	return nil
}

func toOrderResponse(o *domain.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:        o.ID,
		UserID:    o.UserID,
		Amount:    o.Amount,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
	}
}
