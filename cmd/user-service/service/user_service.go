package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go_example/cmd/user-service/domain"
	"go_example/cmd/user-service/dto"
	"go_example/cmd/user-service/repository"
)

var ErrUserNotFound = errors.New("user not found")

// UserService implements user business logic.
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser creates a new user.
func (s *UserService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	u := &domain.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Balance:   req.InitialBalance,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return toUserResponse(u), nil
}

// GetByID returns a user by ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toUserResponse(u), nil
}

// ReserveCredit deducts amount from balance. Returns true if successful.
func (s *UserService) ReserveCredit(ctx context.Context, userID uuid.UUID, amount int64) (bool, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}
	if u.Balance < amount {
		return false, nil
	}
	return true, s.repo.UpdateBalance(ctx, userID, u.Balance-amount)
}

// ReleaseCredit restores amount to balance (compensation).
func (s *UserService) ReleaseCredit(ctx context.Context, userID uuid.UUID, amount int64) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}
	return s.repo.UpdateBalance(ctx, userID, u.Balance+amount)
}

func toUserResponse(u *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Balance:   u.Balance,
		CreatedAt: u.CreatedAt,
	}
}
