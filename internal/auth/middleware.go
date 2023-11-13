package auth

import (
	"context"
	"net/http"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		// Проверка наличия токена
		if tokenString == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		token, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Добавление информации о пользователе в контекст запроса
		ctx := context.WithValue(r.Context(), "userInfo", token.Claims.(*Claims))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
