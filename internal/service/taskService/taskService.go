package taskService

import (
	"context"
	"errors"
	"fmt"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskService struct {
	db                 *postgres.Postgres
	taskRepository     taskRepository
	loaderRepository   loaderRepository
	customerRepository customerRepository
}

type customerRepository interface {
	GetCustomerByIDForUpdate(ctx context.Context, customerID uuid.UUID, tx pgx.Tx) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, customer *models.Customer, tx pgx.Tx) error
}

type loaderRepository interface {
	GetLoadersByIDsForUpdate(ctx context.Context, loaderIDs []uuid.UUID, tx pgx.Tx) ([]models.Loader, error)
	UpdateLoaders(ctx context.Context, loaders []models.Loader, tx pgx.Tx) error
}

type taskRepository interface {
	GetTaskByIDForUpdate(ctx context.Context, taskID uuid.UUID, tx pgx.Tx) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task, tx pgx.Tx) error
	CreateTaskLoaders(ctx context.Context, taskID uuid.UUID, loaderIDs []uuid.UUID, tx pgx.Tx) error
}

func NewTaskService(db *postgres.Postgres, taskRepository taskRepository, loaderRepository loaderRepository, customerRepository customerRepository) *TaskService {
	return &TaskService{db: db, taskRepository: taskRepository, loaderRepository: loaderRepository, customerRepository: customerRepository}
}

func (s *TaskService) StartTask(ctx context.Context, req *models.StartTaskRequest) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	//валидация
	if req.User.UserType != "customer" {
		return errors.New("not customer")
	}

	task, err := s.taskRepository.GetTaskByIDForUpdate(ctx, req.TaskID, tx)
	if err != nil {
		return err
	}
	if task.Status {
		return errors.New("уже выполнена")
	}
	if task.CustomerID != req.User.UserID {
		return errors.New("task.CustomerID != req.User.UserID")
	}
	customer, err := s.customerRepository.GetCustomerByIDForUpdate(ctx, req.User.UserID, tx)
	if err != nil {
		return err
	}
	loaders, err := s.loaderRepository.GetLoadersByIDsForUpdate(ctx, req.LoaderIDs, tx)
	if err != nil {
		return err
	}

	var sumWeightLoaders int
	var sumSalaryLoaders int
	for i := range loaders {
		sumSalaryLoaders += loaders[i].Salary
		if loaders[i].Fatigue == 100 {
			continue
		}
		sumWeightLoaders += loaders[i].MaxWeight * (100 - loaders[i].Fatigue) / 100
		if loaders[i].Drunk {
			loaders[i].Fatigue += 50
		} else {
			loaders[i].Fatigue += 20
		}
		if loaders[i].Fatigue > 100 {
			loaders[i].Fatigue = 100
		}
	}
	if sumWeightLoaders < task.Weight {
		return errors.New("sumWeightLoaders < task.Weight")
	}
	if customer.Capital < sumSalaryLoaders {
		return errors.New(" customer.Capital < sumSalaryLoaders")
	}

	customer.Capital -= sumSalaryLoaders
	task.Status = true

	err = s.loaderRepository.UpdateLoaders(ctx, loaders, tx)
	if err != nil {
		return err
	}

	err = s.customerRepository.UpdateCustomer(ctx, customer, tx)
	if err != nil {
		return err
	}

	err = s.taskRepository.UpdateTask(ctx, task, tx)
	if err != nil {
		return err
	}

	err = s.taskRepository.CreateTaskLoaders(ctx, task.TaskID, req.LoaderIDs, tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx) // Завершаем транзакцию после всех операций
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
