package userService

import (
	"context"
	"errors"
	"fmt"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

type UserService struct {
	db                 *postgres.Postgres
	userRepository     userRepository
	customerRepository customerRepository
	loaderRepository   loaderRepository
	taskRepository     taskRepository
}

type userRepository interface {
	CreateUser(ctx context.Context, user *models.User, tx pgx.Tx) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UsernameExists(ctx context.Context, username string, tx pgx.Tx) (bool, error)
}

type customerRepository interface {
	CreateCustomer(ctx context.Context, customer *models.Customer, tx pgx.Tx) error
	HasCustomers(ctx context.Context, tx pgx.Tx) (bool, error)
	GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error)
}

type loaderRepository interface {
	CreateLoader(ctx context.Context, loader *models.Loader, tx pgx.Tx) error
	GetLoaderByID(ctx context.Context, loaderID uuid.UUID) (*models.Loader, error)
	GetLoaders(ctx context.Context) ([]models.Loader, error)
}

type taskRepository interface {
	CreateTasks(ctx context.Context, tasks []models.Task, tx pgx.Tx) error
	GetTasksCustomers(ctx context.Context, customerID uuid.UUID) ([]models.Task, error)
	GetTasksLoaders(ctx context.Context, loaderID uuid.UUID) ([]models.Task, error)
}

func NewUserService(db *postgres.Postgres, userRepository userRepository, customerRepository customerRepository, loaderRepository loaderRepository, taskRepository taskRepository) *UserService {
	return &UserService{db: db, userRepository: userRepository, customerRepository: customerRepository, loaderRepository: loaderRepository, taskRepository: taskRepository}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	exist, err := s.userRepository.UsernameExists(ctx, user.Username, tx)
	if err != nil {
		return nil, err
	}
	if exist == true {
		return nil, errors.New("username exist")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPassword)

	user, err = s.userRepository.CreateUser(ctx, user, tx)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())

	switch user.UserType {
	case "customer":
		exist, err := s.customerRepository.HasCustomers(ctx, tx)
		if err != nil {
			return nil, err
		}
		if exist == true {
			return nil, errors.New("exist")
		}

		customer := &models.Customer{
			CustomerID: user.UserID,
			Capital:    rand.Intn(100000-10000+1) + 10000,
		}
		err = s.customerRepository.CreateCustomer(ctx, customer, tx)
		if err != nil {
			return nil, err
		}

		taskCount := rand.Intn(5) + 1
		tasks := make([]models.Task, 0, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks = append(tasks, models.Task{
				CustomerID:  user.UserID,
				Weight:      rand.Intn(80-10+1) + 10,
				Description: "",
				Status:      false,
			})
		}
		err = s.taskRepository.CreateTasks(ctx, tasks, tx)
		if err != nil {
			return nil, err
		}

	case "loader":
		var drunk bool
		if rand.Intn(2) == 0 {
			drunk = true
		}
		loader := &models.Loader{
			LoaderID:  user.UserID,
			MaxWeight: rand.Intn(30-5+1) + 5,
			Drunk:     drunk,
			Fatigue:   rand.Intn(101),
			Salary:    rand.Intn(30000-10000+1) + 10000,
		}
		err = s.loaderRepository.CreateLoader(ctx, loader, tx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(ctx) // Завершаем транзакцию после всех операций
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return user, nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, credentials *models.User) (*models.User, error) {
	user, err := s.userRepository.GetUserByUsername(ctx, credentials.Username)
	if err != nil {
		return nil, err // Пользователь не найден или другая ошибка
	}

	// Сравнение хешированного пароля с паролем из запроса
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		return nil, errors.New("invalid credentials") // Неверный пароль
	}

	return user, nil // Успешная аутентификация
}

func (s *UserService) GetUserDetails(ctx context.Context, user *models.User) (interface{}, error) {
	switch user.UserType {
	case "customer":
		var customerResponce struct {
			Info    *models.Customer `json:"info"`
			Loaders []models.Loader  `json:"loaders"`
		}
		info, err := s.customerRepository.GetCustomerByID(ctx, user.UserID)
		if err != nil {
			return nil, err
		}
		customerResponce.Info = info

		loaders, err := s.loaderRepository.GetLoaders(ctx)
		if err != nil {
			return nil, err
		}
		customerResponce.Loaders = loaders
		return customerResponce, nil
	case "loader":
		return s.loaderRepository.GetLoaderByID(ctx, user.UserID)
	default:
		return nil, errors.New("err")
	}
}

func (s *UserService) GetUserTasks(ctx context.Context, user *models.User) (interface{}, error) {
	switch user.UserType {
	case "customer":
		return s.taskRepository.GetTasksCustomers(ctx, user.UserID)
	case "loader":
		return s.taskRepository.GetTasksLoaders(ctx, user.UserID)
	default:
		return nil, errors.New("err")
	}
}
