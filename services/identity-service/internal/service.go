package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, eventData []byte) error
}

type Service struct {
	repo      UserRepository
	publisher EventPublisher
}

func NewService(r UserRepository, p EventPublisher) *Service {
	return &Service{
		repo:      r,
		publisher: p,
	}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (*User, error) {
	if _, err := s.repo.FindByEmail(ctx, email); !errors.Is(err, ErrUserNotFound) {
		if err == nil {
			return nil, ErrEmailAlreadyInUse
		}
		return nil, fmt.Errorf("check user by email for register: %v", err)
	}

	salt := []byte("somesalt")
	hash := argon2.IDKey([]byte(password), salt, 1, 10*1024, 4, 32)
	passwordHash := fmt.Sprintf("%x", hash) // store as hex

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
