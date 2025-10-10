package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	tokenRepo    TokenRepository
	publisher    EventPublisher
	tokenManager *JWTManager
	passManager  *PasswordManager
}

func NewService(
	r UserRepository,
	tk TokenRepository,
	tm *JWTManager,
	pm *PasswordManager,
	p EventPublisher,
) *Service {
	return &Service{
		repo:         r,
		tokenRepo:    tk,
		tokenManager: tm,
		publisher:    p,
		passManager:  pm,
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

func (s *Service) Login(ctx context.Context, email, password string) (accesToken, refreshToken string, err error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("authentication failed: %w", ErrUserNotFound)
	}

	match, err := s.passManager.Verify(password, user.PasswordHash)
	if err != nil || !match {
		return "", "", fmt.Errorf("authentication failed")
	}

	accessToken, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %v", err)
	}

	refreshToken, refreshTokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %v", err)
	}

	refreshTokenTTL := time.Hour * 24 * 7 // 7 days
	rt := &RefreshToken{
		TokenHash: refreshTokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(refreshTokenTTL).Unix(),
	}

	// for now, lets assume the token repo exists
	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %v", err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	userID, err := s.tokenRepo.Revoke(ctx, tokenHash)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %v", err)
	}

	newAccessToken, err := s.tokenManager.Generate(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %v", err)
	}

	newRefreshToken, newRefreshTokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %v", err)
	}

	refreshTokenTTL := time.Hour * 24 * 7
	rt := &RefreshToken{
		TokenHash: newRefreshTokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(refreshTokenTTL).Unix(),
	}

	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %v", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.tokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke tokens on logout: %v", err)
	}

	return nil
}
