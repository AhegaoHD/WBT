package repository

import (
	"context"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *postgres.Postgres
}

func NewUserRepository(db *postgres.Postgres) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User, tx pgx.Tx) (*models.User, error) {
	const query = `
		INSERT INTO users (username, password, user_type)
		VALUES ($1, $2, $3)
		RETURNING user_id`

	err := tx.QueryRow(ctx, query, user.Username, user.Password, user.UserType).Scan(&user.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string, tx pgx.Tx) (bool, error) {
	const query = `SELECT COUNT(*) FROM users WHERE username = $1`

	var count int
	err := tx.QueryRow(ctx, query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const query = `
		SELECT user_id, username, password, user_type
		FROM users
		WHERE username = $1`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(&user.UserID, &user.Username, &user.Password, &user.UserType)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
