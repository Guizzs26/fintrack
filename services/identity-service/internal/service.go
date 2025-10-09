package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	repo         UserRepository
	tokenRepo    TokenRepository
	publisher    EventPublisher
	tokenManager *JWTManager
}

func NewService(
	r UserRepository,
	tk TokenRepository,
	tm *JWTManager,
	p EventPublisher,
) *Service {
	return &Service{
		repo:         r,
		tokenRepo:    tk,
		tokenManager: tm,
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

func (s *Service) Login(ctx context.Context, email, password string) (accesToken, refreshToken string, err error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("authentication failed: %w", ErrUserNotFound)
	}

	if err := user.ComparePassword(password); err != nil {
		return "", "", fmt.Errorf("authentication failed: %v", err)
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
		ExpiresAt: time.Now().Add(refreshTokenTTL),
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

	if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %v", err)
	}

	// TODO: Before generating a new one, we should get the UserID of the revoked token
	// to know who to issue the new one to. The current Revoke implementation doesn't return the UserID.
	// For now, let's assume we have the UserID.
	var userID uuid.UUID

	newAccessToken, err := s.tokenManager.Generate(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %v", err)
	}

	newRefreshToken, newRefreshTokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %v", err)
	}

	rt := &RefreshToken{
		TokenHash: newRefreshTokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %v", err)
	}

	return newAccessToken, newRefreshToken, nil
}
