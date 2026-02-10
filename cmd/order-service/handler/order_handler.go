package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"go_example/cmd/order-service/dto"
	"go_example/cmd/order-service/service"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrder creates a new order (starts saga). POST /orders
func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	var req dto.CreateOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Amount < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "amount must be positive"})
	}
	order, err := h.svc.CreateOrder(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

// ListByUserID returns orders for a user (used by user-service). GET /orders?userId=xxx
func (h *OrderHandler) ListByUserID(c fiber.Ctx) error {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "userId query parameter is required"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid userId"})
	}
	orders, err := h.svc.ListByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if orders == nil {
		orders = []*dto.OrderResponse{}
	}
	return c.JSON(orders)
}

// GetByID returns an order by ID. GET /orders/:id
func (h *OrderHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	order, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if err == service.ErrOrderNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(order)
}

// CancelOrder cancels an order (triggers compensation). DELETE /orders/:id
func (h *OrderHandler) CancelOrder(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	if err := h.svc.CancelOrder(c.Context(), id); err != nil {
		if err == service.ErrOrderNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
