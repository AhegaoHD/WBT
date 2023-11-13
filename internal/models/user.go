package models

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Password string    `json:"password"` // Храните хешированный пароль
	UserType string    `json:"user_type"`
}

type Customer struct {
	CustomerID uuid.UUID `json:"customer_id"`
	Capital    int       `json:"capital"`
}

type Loader struct {
	LoaderID  uuid.UUID `json:"loader_id"`
	MaxWeight int       `json:"max_weight"`
	Drunk     bool      `json:"drunk"`
	Fatigue   int       `json:"fatigue"`
	Salary    int       `json:"salary"`
}
