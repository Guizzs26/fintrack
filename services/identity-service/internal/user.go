package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var (
	ErrEmailAlreadyInUse = errors.New("email is already in use")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type TokenRepository interface {
	Save(ctx context.Context, token *RefreshToken) error
	Revoke(ctx context.Context, tokenHash string) error
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
	ExpiresAt time.Time `dynamodbav:"ExpiresAt"`
}

func (u *User) ComparePassword(password string) error {
	salt := []byte("somesalt")
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	passwordHash := fmt.Sprintf("%x", hash)

	if passwordHash != u.PasswordHash {
		return errors.New("invalid credentials")
	}
	return nil
}
