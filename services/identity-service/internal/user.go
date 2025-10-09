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
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
