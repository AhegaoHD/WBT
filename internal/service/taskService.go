package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
)

type TaskService struct {
	db                 *postgres.Postgres
	taskRepository     taskRepository
	loaderRepository   loaderRepository
	userRepository     userRepository
	customerRepository customerRepository
}

func NewTaskService(db *postgres.Postgres, taskRepository taskRepository, loaderRepository loaderRepository, userRepository userRepository, customerRepository customerRepository) *TaskService {
	return &TaskService{db: db, taskRepository: taskRepository, loaderRepository: loaderRepository, userRepository: userRepository, customerRepository: customerRepository}
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
