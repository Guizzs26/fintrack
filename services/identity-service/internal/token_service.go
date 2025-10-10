package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ TokenManager = (*TokenService)(nil)

type TokenService struct {
	tokenRepo       TokenRepository
	jwtGenerator    TokenGenerator
	refreshTokenTTL time.Duration
}

func NewTokenService(repo TokenRepository, jwtGen TokenGenerator, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		tokenRepo:       repo,
		jwtGenerator:    jwtGen,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *TokenService) NewPairForUser(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	accessToken, err := s.jwtGenerator.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	refreshToken, refreshTokenHash, err := s.generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	rt := &RefreshToken{
		TokenHash: refreshTokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL).Unix(),
	}
	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %v", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *TokenService) RotateRefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Revoke the old token. Successful revocation proves the token was valid
	// and returns the UserID it belonged to
	userID, err := s.tokenRepo.Revoke(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token: %v", err)
	}

	return s.NewPairForUser(ctx, userID)
}

func (s *TokenService) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.RevokeAllForUser(ctx, userID)
}

func (s *TokenService) generateOpaqueToken() (token, hash string, err error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(randomBytes)

	hashBytes := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(hashBytes[:])
	return token, hash, nil
}
