package middleware

import (
	"context"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

type JWTMiddleware struct {
	jwtService jwtService
}

type jwtService interface {
	ValidateToken(tokenString string) (*jwt.Token, error)
}

func NewJWTMiddleware(jwtService jwtService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		// Проверка наличия токена
		if tokenString == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		token, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.Claims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		user := claims.User
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
