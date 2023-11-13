package authController

import (
	"context"
	"encoding/json"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/gorilla/mux"
	"net/http"
)

type AuthController struct {
	userService userService
	jwtService  jwtService
}

type userService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	AuthenticateUser(ctx context.Context, credentials *models.User) (*models.User, error)
}

type jwtService interface {
	GenerateToken(user *models.User) (string, error)
}

func NewAuthController(userService userService, jwtService jwtService) *AuthController {
	return &AuthController{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (c *AuthController) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/login", c.LoginHandler).Methods("POST")
	r.HandleFunc("/register", c.RegisterHandler).Methods("POST")
}

func (c *AuthController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user *models.User

	// Декодирование запроса
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = validateRegister(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err = c.userService.CreateUser(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	token, err := c.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Отправка ответа с JWT токеном
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (c *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials *models.User

	// Декодирование запроса
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := c.userService.AuthenticateUser(r.Context(), credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	token, err := c.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Отправка ответа с JWT токеном
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
