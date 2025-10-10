package identity

import (
	"context"

	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenGenerator interface {
	Generate(userID uuid.UUID) (string, error)
}

type TokenManager interface {
	NewPairForUser(ctx context.Context, userID uuid.UUID) (*TokenPair, error)
	RotateRefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}
