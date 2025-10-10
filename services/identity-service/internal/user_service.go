package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, eventData []byte) error
}

type Service struct {
	repo         UserRepository
	tokenManager TokenManager
	passManager  *PasswordManager
	publisher    EventPublisher
}

func NewService(
	r UserRepository,
	tm TokenManager,
	pm *PasswordManager,
	p EventPublisher,
) *Service {
	return &Service{
		repo:         r,
		tokenManager: tm,
		passManager:  pm,
		publisher:    p,
	}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (*User, error) {
	if _, err := s.repo.FindByEmail(ctx, email); !errors.Is(err, ErrUserNotFound) {
		if err == nil {
			return nil, ErrEmailAlreadyInUse
		}
		return nil, fmt.Errorf("check user by email for register: %v", err)
	}

	passwordHash, err := s.passManager.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save user in register: %v", err)
	}

	// TODO -> Publish event in kafka

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", ErrUserNotFound)
	}

	match, err := s.passManager.Verify(password, user.PasswordHash)
	if err != nil || !match {
		return nil, fmt.Errorf("authentication failed")
	}

	return s.tokenManager.NewPairForUser(ctx, user.ID)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	pair, err := s.tokenManager.RotateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.tokenManager.RevokeAllForUser(ctx, userID)
}
