package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Payload JWT-токена.
type Claims struct {
	UserID int64  `json:"uid"`
	Kind   string `json:"kind"` // "access" | "refresh"
	jwt.RegisteredClaims
}

type TokenService interface {
	GeneratePair(userID int64) (access, refresh string, err error)
	ValidateAccess(token string) (int64, error)
	ValidateRefresh(token string) (int64, error)
}

type tokenService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenService(secret string, accessTTL, refreshTTL time.Duration) TokenService {
	return &tokenService{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GeneratePair создаёт пару access + refresh токенов.
func (s *tokenService) GeneratePair(userID int64) (access, refresh string, err error) {
	access, err = s.sign(userID, "access", s.accessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("sign access: %w", err)
	}
	refresh, err = s.sign(userID, "refresh", s.refreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh: %w", err)
	}
	return access, refresh, nil
}

// ValidateAccess проверяет access-токен и возвращает userID.
func (s *tokenService) ValidateAccess(tokenStr string) (int64, error) {
	return s.validate(tokenStr, "access")
}

// ValidateRefresh проверяет refresh-токен и возвращает userID.
func (s *tokenService) ValidateRefresh(tokenStr string) (int64, error) {
	return s.validate(tokenStr, "refresh")
}

func (s *tokenService) sign(userID int64, kind string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Kind:   kind,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *tokenService) validate(tokenStr, expectedKind string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	if claims.Kind != expectedKind {
		return 0, fmt.Errorf("expected %s token, got %s", expectedKind, claims.Kind)
	}
	return claims.UserID, nil
}
