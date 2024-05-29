package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

const JWT_EXPIRES_IN = 24 * time.Hour

func newJWT(secret string) (string, error) {
	userID := uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWT_EXPIRES_IN)),
		},
		UserID: userID,
	})
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func GetUserID(jwtToken, secret string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims,
		func(t *jwt.Token) (interface{}, error) {
			if method := t.Method.Alg(); method != jwt.SigningMethodHS256.Name {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(secret), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		logger.Log.Info("Invalid JWT token")
		return "", errors.New("invalid JWT token")
	}

	return claims.UserID.String(), nil
}
