package repository

import (
	"context"
	"errors"
	"github.com/AhegaoHD/WBT/internal/models"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LoaderRepository struct {
	db *postgres.Postgres
}

func NewLoaderRepository(db *postgres.Postgres) *LoaderRepository {
	return &LoaderRepository{db: db}
}

func (r *LoaderRepository) CreateLoader(ctx context.Context, loader *models.Loader, tx pgx.Tx) error {
	const query = `INSERT INTO loaders (loader_id, max_weight, drunk, fatigue, salary) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.Exec(ctx, query, loader.LoaderID, loader.MaxWeight, loader.Drunk, loader.Fatigue, loader.Salary)
	return err
}

func (r *LoaderRepository) GetLoaderByID(ctx context.Context, loaderID uuid.UUID) (*models.Loader, error) {
	const query = `SELECT loader_id, max_weight, drunk, fatigue, salary FROM loaders WHERE loader_id = $1`
	var loader models.Loader
	err := r.db.Pool.QueryRow(ctx, query, loaderID).Scan(&loader.LoaderID, &loader.MaxWeight, &loader.Drunk, &loader.Fatigue, &loader.Salary)
	if err != nil {
		return nil, err
	}
	return &loader, nil
}

func (r *LoaderRepository) GetLoaders(ctx context.Context) ([]models.Loader, error) {
	const query = `SELECT loader_id, max_weight, drunk, fatigue, salary FROM loaders`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loaders []models.Loader
	for rows.Next() {
		var loader models.Loader
		err = rows.Scan(&loader.LoaderID, &loader.MaxWeight, &loader.Drunk, &loader.Fatigue, &loader.Salary)
		if err != nil {
			return nil, err
		}
		loaders = append(loaders, loader)
	}

	// Проверка на ошибки, возникшие во время итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return loaders, nil
}

func (r *LoaderRepository) GetLoadersByIDsForUpdate(ctx context.Context, loaderIDs []uuid.UUID, tx pgx.Tx) ([]models.Loader, error) {

	const query = `SELECT loader_id, max_weight, drunk, fatigue, salary FROM loaders WHERE loader_id = ANY($1) FOR UPDATE `

	rows, err := tx.Query(ctx, query, loaderIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loaders []models.Loader
	for rows.Next() {
		var loader models.Loader
		err = rows.Scan(&loader.LoaderID, &loader.MaxWeight, &loader.Drunk, &loader.Fatigue, &loader.Salary)
		if err != nil {
			return nil, err
		}
		loaders = append(loaders, loader)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(loaders) != len(loaderIDs) {
		return nil, errors.New("wrong loaders")
	}

	return loaders, nil
}

func (r *LoaderRepository) UpdateLoaders(ctx context.Context, loaders []models.Loader, tx pgx.Tx) error {
	batch := &pgx.Batch{}

	const query = `UPDATE loaders SET max_weight = $1, drunk = $2, fatigue = $3, salary = $4 WHERE loader_id = $5`
	for _, loader := range loaders {
		batch.Queue(query, loader.MaxWeight, loader.Drunk, loader.Fatigue, loader.Salary, loader.LoaderID)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Проверяем результаты выполнения каждого запроса в пакете
	for range loaders {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
