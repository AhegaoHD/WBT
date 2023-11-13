package repository

import (
	"context"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CustomerRepository struct {
	db *postgres.Postgres
}

func NewCustomerRepository(db *postgres.Postgres) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) CreateCustomer(ctx context.Context, customer *models.Customer, tx pgx.Tx) error {
	const query = `INSERT INTO customers (customer_id, capital) VALUES ($1, $2)`
	_, err := tx.Exec(ctx, query, customer.CustomerID, customer.Capital)
	return err
}

func (r *CustomerRepository) GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error) {
	const query = `SELECT customer_id, capital FROM customers WHERE customer_id = $1`
	var customer models.Customer
	err := r.db.Pool.QueryRow(ctx, query, customerID).Scan(&customer.CustomerID, &customer.Capital)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) HasCustomers(ctx context.Context, tx pgx.Tx) (bool, error) {
	const query = `SELECT COUNT(*) FROM customers`
	var count int
	err := tx.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CustomerRepository) GetCustomerByIDForUpdate(ctx context.Context, customerID uuid.UUID, tx pgx.Tx) (*models.Customer, error) {
	const query = `SELECT customer_id, capital FROM customers WHERE customer_id = $1 FOR UPDATE `
	var customer models.Customer
	err := tx.QueryRow(ctx, query, customerID).Scan(&customer.CustomerID, &customer.Capital)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) UpdateCustomer(ctx context.Context, customer *models.Customer, tx pgx.Tx) error {
	const query = `UPDATE customers SET capital = $1 WHERE customer_id = $2`

	_, err := tx.Exec(ctx, query, customer.Capital, customer.CustomerID)
	if err != nil {
		return err
	}

	return nil
}
