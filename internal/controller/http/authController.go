package httpController

import (
	"context"
	"encoding/json"
	"github.com/AhegaoHD/WBT/internal/auth"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type AuthController struct {
	userService userService
}

type userService interface {
	CreateUser(ctx context.Context, user *models.User) error
	AuthenticateUser(ctx context.Context, credentials *models.User) (*models.User, error)
	GetUserDetails(ctx context.Context, user *models.User) (interface{}, error)
	GetUserTasks(ctx context.Context, user *models.User) (interface{}, error)
}

func NewAuthController(userService userService) *AuthController {
	return &AuthController{
		userService: userService,
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

	err = c.userService.CreateUser(r.Context(), user)
	if err != nil {
		log.Println(err)
		http.Error(w, "Err", http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Отправка ответа с JWT токеном
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (c *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials models.User

	// Декодирование запроса
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := c.userService.AuthenticateUser(r.Context(), &credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Отправка ответа с JWT токеном
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
