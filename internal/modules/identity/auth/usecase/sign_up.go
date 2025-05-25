package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Guizzs26/fintrack/internal/modules/identity/auth/infra/hasher"
	"github.com/Guizzs26/fintrack/internal/modules/identity/auth/infra/persistence"
	"github.com/google/uuid"
)

type SignUpInput struct {
	Name     string
	Email    string
	Password string
}

type SignUpOutput struct {
	UserID string
	// Token string
	Name  string
	Email string
}

type SignUpUseCase struct {
	repo   persistence.AuthRepository
	hasher hasher.Hasher
}

func NewSignUpUseCase(repo persistence.AuthRepository, hasher hasher.Hasher) *SignUpUseCase {
	return &SignUpUseCase{repo: repo, hasher: hasher}
}

func (uc *SignUpUseCase) Execute(ctx context.Context, input SignUpInput) (*SignUpOutput, error) {
	existingEmail, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, fmt.Errorf("email already in use")
	}

	hashedPass, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	userID := uuid.New().String()
	auth := &persistence.AuthDB{
		ID:        userID,
		Email:     input.Email,
		Password:  hashedPass,
		CreatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, auth); err != nil {
		return nil, err
	}

	return &SignUpOutput{
		UserID: userID,
		Name:   input.Name,
		Email:  input.Email,
	}, nil
}
