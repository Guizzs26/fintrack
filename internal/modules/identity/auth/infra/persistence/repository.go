package persistence

import "context"

type AuthRepository interface {
	Create(ctx context.Context, auth *AuthDB) error
	FindByEmail(ctx context.Context, email string) (*AuthDB, error)
}
