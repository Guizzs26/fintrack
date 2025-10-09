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
}

type User struct {
	ID           uuid.UUID `dynamodbav:"ID"`
	Name         string    `dynamodbav:"Name"`
	Email        string    `dynamodbav:"Email"`
	PasswordHash string    `dynamodbav:"PasswordHash"`
	CreatedAt    time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time `dynamodbav:"UpdatedAt"`
}
