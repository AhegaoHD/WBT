package jwtService

import (
	"errors"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JWTService struct {
	secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: secret,
	}
}

func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &models.Claims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))

	return tokenString, err
}

func (s *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
