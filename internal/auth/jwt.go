package auth

import (
	"errors"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

var jwtKey = []byte("secret_key")

func GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		UserID:   user.UserID,
		Username: user.Username,
		UserType: user.UserType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	return tokenString, err
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	UserType string    `json:"user_type"`
	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
