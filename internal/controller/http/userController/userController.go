package userController

import (
	"context"
	"encoding/json"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type UsersController struct {
	userService userService
	taskService taskService
	middleware  middleware
}

type userService interface {
	GetUserDetails(ctx context.Context, user *models.User) (interface{}, error)
	GetUserTasks(ctx context.Context, user *models.User) (interface{}, error)
}

type taskService interface {
	StartTask(ctx context.Context, req *models.StartTaskRequest) error
}

type middleware interface {
	Middleware(next http.Handler) http.Handler
}

func NewUsersController(userService userService, taskService taskService, middleware middleware) *UsersController {
	return &UsersController{userService: userService, taskService: taskService, middleware: middleware}
}

func (c *UsersController) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("").Subrouter()
	api.Use(c.middleware.Middleware)
	api.HandleFunc("/me", c.GetUserDetails).Methods("GET")
	api.HandleFunc("/tasks", c.GetUserTasks).Methods("GET")
	api.HandleFunc("/start", c.StartTask).Methods("POST")
}

func (c *UsersController) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userDetails, err := c.userService.GetUserDetails(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.writeJSONResponse(w, http.StatusOK, userDetails)
}

func (c *UsersController) GetUserTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userDetails, err := c.userService.GetUserTasks(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.writeJSONResponse(w, http.StatusOK, userDetails)
}

func (c *UsersController) StartTask(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var startTask *models.StartTaskRequest

	// Декодирование запроса
	err := json.NewDecoder(r.Body).Decode(&startTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	startTask.User = user

	err = validateStartTask(startTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.taskService.StartTask(r.Context(), startTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *UsersController) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Failed to encode response:", err)
	}
}
