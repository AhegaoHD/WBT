package models

import "github.com/google/uuid"

type Task struct {
	TaskID      uuid.UUID `json:"task_id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	Weight      int       `json:"weight"`
	Description string    `json:"description"`
	Status      bool      `json:"status"`
}

type TaskLoader struct {
	TaskID   uuid.UUID `json:"task_id"`
	LoaderID uuid.UUID `json:"loader_id"`
}

type StartTaskRequest struct {
	User      *User
	TaskID    uuid.UUID   `json:"task_id"`
	LoaderIDs []uuid.UUID `json:"loader_ids"`
}
