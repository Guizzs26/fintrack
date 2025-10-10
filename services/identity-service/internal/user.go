package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyInUse = errors.New("email is already in use")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type TokenRepository interface {
	Save(ctx context.Context, token *RefreshToken) error
	Revoke(ctx context.Context, tokenHash string) (uuid.UUID, error)
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type User struct {
	ID           uuid.UUID `dynamodbav:"ID"`
	Name         string    `dynamodbav:"Name"`
	Email        string    `dynamodbav:"Email"`
	PasswordHash string    `dynamodbav:"PasswordHash"`
	CreatedAt    time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time `dynamodbav:"UpdatedAt"`
}

type RefreshToken struct {
	TokenHash string    `dynamodbav:"TokenHash"`
	UserID    uuid.UUID `dynamodbav:"UserID"`
	ExpiresAt int64     `dynamodbav:"ExpiresAt"`
}
