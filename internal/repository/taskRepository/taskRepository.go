package taskRepository

import (
	"context"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	db *postgres.Postgres
}

func NewTaskRepository(db *postgres.Postgres) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTasks(ctx context.Context, tasks []models.Task, tx pgx.Tx) error {
	const query = `INSERT INTO tasks (customer_id, weight, description, status) VALUES ($1, $2, $3, $4)`

	batch := &pgx.Batch{}
	for _, task := range tasks {
		batch.Queue(query, task.CustomerID, task.Weight, task.Description, task.Status)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Проверяем результаты выполнения каждого запроса в пакете
	for range tasks {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TaskRepository) GetTasksCustomers(ctx context.Context, customerID uuid.UUID) ([]models.Task, error) {
	const query = `SELECT task_id, customer_id, weight, description, status FROM tasks WHERE customer_id = $1 AND status = '0'`

	rows, err := r.db.Pool.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err = rows.Scan(&task.TaskID, &task.CustomerID, &task.Weight, &task.Description, &task.Status)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) GetTasksLoaders(ctx context.Context, loaderID uuid.UUID) ([]models.Task, error) {
	const query = `SELECT t.task_id, t.customer_id, t.weight, t.description, t.status 
                   FROM tasks t 
                   JOIN task_loaders tl ON t.task_id = tl.task_id 
                   WHERE tl.loader_id = $1`

	rows, err := r.db.Pool.Query(ctx, query, loaderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err = rows.Scan(&task.TaskID, &task.CustomerID, &task.Weight, &task.Description, &task.Status)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) GetTaskByIDForUpdate(ctx context.Context, taskID uuid.UUID, tx pgx.Tx) (*models.Task, error) {
	const query = `SELECT task_id, customer_id, weight, description, status FROM tasks WHERE task_id = $1 FOR UPDATE`

	var task models.Task
	err := tx.QueryRow(ctx, query, taskID).Scan(&task.TaskID, &task.CustomerID, &task.Weight, &task.Description, &task.Status)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, task *models.Task, tx pgx.Tx) error {
	const query = `UPDATE tasks SET customer_id = $1, weight = $2, description = $3, status = $4 WHERE task_id = $5`

	_, err := tx.Exec(ctx, query, task.CustomerID, task.Weight, task.Description, task.Status, task.TaskID)
	if err != nil {
		return err
	}

	return nil
}

//type TaskLoaderRepository struct {
//	db *postgres.Postgres
//}
//
//func NewTaskLoaderRepository(db *postgres.Postgres) *TaskLoaderRepository {
//	return &TaskLoaderRepository{db: db}
//}

func (r *TaskRepository) CreateTaskLoaders(ctx context.Context, taskID uuid.UUID, loaderIDs []uuid.UUID, tx pgx.Tx) error {
	batch := &pgx.Batch{}

	const query = `INSERT INTO task_loaders (task_id, loader_id) VALUES ($1, $2)`
	for _, loaderID := range loaderIDs {
		batch.Queue(query, taskID, loaderID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Проверяем результаты выполнения каждого запроса в пакете
	for range loaderIDs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
