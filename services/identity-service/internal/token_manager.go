package identity

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secretKey      []byte
	accessTokenTTL time.Duration
}

func NewJWTManager(sk string, attl time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:      []byte(sk),
		accessTokenTTL: attl,
	}
}

func (m *JWTManager) Generate(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(m.accessTokenTTL).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}
