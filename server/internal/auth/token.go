// Package auth — autentifikatsiya: mehmon + akkaunt (parol) + JWT.
// Telegram provider keyin (Bosqich 2) qo'shiladi.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — JWT yuki.
type Claims struct {
	Role    string `json:"role"`
	IsGuest bool   `json:"guest"`
	jwt.RegisteredClaims
}

// TokenManager — JWT chiqarish/tekshirish (HS256).
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) Issue(userID, role string, guest bool) (string, error) {
	now := time.Now()
	claims := Claims{
		Role:    role,
		IsGuest: guest,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *TokenManager) Verify(tokenStr string) (*Claims, error) {
	var claims Claims
	tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("kutilmagan imzo usuli")
		}
		return m.secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("token yaroqsiz")
	}
	return &claims, nil
}
