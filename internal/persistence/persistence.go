package persistence

import (
	"context"
	"database/sql"

	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/cars"

	"github.com/jmoiron/sqlx"
	// pq is imported fore the postgres drivers.
	_ "github.com/lib/pq"
)

// signed file to generate mock
//go:generate mockgen -source ./persistence.go -destination mocks/persistence.mock.go -package mocks

// Repository persistence methods.
//
//nolint:interfacebloat
type Repository interface {
	GetAllCars(ctx context.Context, cityID string, category string, startKey uint, count uint) ([]models.Cars, error)
	UpdateCar(ctx context.Context, updatePayLoad models.Cars, carID string) (*models.Cars, error)
	RegisterCar(ctx context.Context, carPayload models.Cars) (*models.Cars, error)
	GetCarsByID(ctx context.Context, carID string) (*models.Cars, error)
	PlaceBid(ctx context.Context, bid models.Bids) (*models.Bids, error)
	GetBidByID(ctx context.Context, bidID string) (*models.Bids, error)
}

// carRow contains the columns for an event.
type carRow struct {
	ID         string       `db:"id"`
	Properties *models.Cars `db:"properties"`
}

// RepositoryPg is a postgres implementation of Repository.
type RepositoryPg struct {
	db *sqlx.DB
}

// This line ensures that the RepositoryPg struct implements the Repository interface.
var _ Repository = &RepositoryPg{}

func NewRepository(db *sql.DB) (*RepositoryPg, error) {
	pgDB := sqlx.NewDb(db, "postgres")

	return &RepositoryPg{
		db: pgDB,
	}, nil
}

// GetAllEvents implements Repository.
func (r *RepositoryPg) GetAllCars(ctx context.Context, cityID string, category string, startKey uint, count uint) ([]models.Cars, error) {
	rows := []carRow{}

	err := r.db.SelectContext(ctx, &rows, `SELECT id, properties FROM cars WHERE ($2 = '' OR properties->>'category' = $2) ORDER BY properties->>'city_id' = $1 DESC, properties->>'date' ASC LIMIT $4 OFFSET $3`,
		cityID, category, startKey, count)
	if err != nil {
		return nil, err
	}

	carSlice := make([]models.Cars, len(rows))

	for i := range rows {
		carSlice[i] = *rows[i].Properties
		carSlice[i].ID = rows[i].ID
	}

	return carSlice, nil
}

func (r *RepositoryPg) RegisterCar(ctx context.Context, carPayload models.Cars) (*models.Cars, error) {
	row := carRow{}
	err := r.db.GetContext(ctx, &row, `INSERT INTO cars(properties) VALUES($1) RETURNING id, properties`, carPayload)
	if err != nil {
		return nil, err
	}
	row.Properties.ID = row.ID

	return row.Properties, nil
}

func (r *RepositoryPg) GetCarsByID(ctx context.Context, carID string) (*models.Cars, error) {
	row := carRow{}
	err := r.db.GetContext(ctx, &row, "SELECT id, properties FROM cars WHERE id = $1", carID)

	if err != nil {
		return nil, err
	}
	row.Properties.ID = row.ID

	return row.Properties, nil
}

func (r *RepositoryPg) UpdateCar(ctx context.Context, updatePayLoad models.Cars, carID string) (*models.Cars, error) {
	row := carRow{}
	err := r.db.GetContext(ctx, &row, "UPDATE cars SET properties=$1 WHERE id = $2 RETURNING id, properties", updatePayLoad, carID)

	if err != nil {
		return nil, err
	}
	row.Properties.ID = row.ID

	return row.Properties, nil
}

func (r *RepositoryPg) PlaceBid(ctx context.Context, bid models.Bids) (*models.Bids, error) {
	createdBid := models.Bids{}
	err := r.db.GetContext(ctx, &createdBid, `INSERT INTO bids(car_id, bid_amount,email,user_name) VALUES($1,$2,$3,$4)RETURNING bid_id, car_id, bid_amount,email,user_name, created_at`, bid.CarID, bid.Amount, bid.Email, bid.UserName)

	if err != nil {
		return nil, err
	}

	return &createdBid, nil
}

func (r *RepositoryPg) GetBidByID(ctx context.Context, bidID string) (*models.Bids, error) {
	bids := models.Bids{}
	err := r.db.GetContext(ctx, &bids, `SELECT * FROM bids WHERE bid_id=$1`, bidID)
	if err != nil {
		return nil, err
	}
	return &bids, nil
}
