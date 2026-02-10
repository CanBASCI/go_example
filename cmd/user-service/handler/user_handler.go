package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"go_example/cmd/user-service/dto"
	"go_example/cmd/user-service/orderclient"
	"go_example/cmd/user-service/service"
)

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	svc            *service.UserService
	orderServiceURL string
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService, orderServiceURL string) *UserHandler {
	return &UserHandler{svc: svc, orderServiceURL: orderServiceURL}
}

// CreateUser creates a new user. POST /users
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username is required"})
	}
	if req.InitialBalance < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "initialBalance must be non-negative"})
	}
	user, err := h.svc.CreateUser(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetByID returns a user by ID. GET /users/:id
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	user, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}

// GetUserWithOrders returns user and their orders (aggregated from user-service and order-service). GET /users/:id/orders
func (h *UserHandler) GetUserWithOrders(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	user, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	orders, err := orderclient.ListByUserID(c.Context(), h.orderServiceURL, id)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch orders", "detail": err.Error()})
	}
	return c.JSON(fiber.Map{"user": user, "orders": orders})
}
