package persistence

import (
	"context"
	"database/sql"

	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/event"

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
	GetEvents(ctx context.Context, cityID string, category string, startKey uint, count uint) ([]models.Event, error)
}

// EventRow contains the columns for an event.
type EventRow struct {
	ID    string        `db:"id"`
	Event *models.Event `db:"event"`
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
func (r *RepositoryPg) GetEvents(ctx context.Context, cityID string, category string, startKey uint, count uint) ([]models.Event, error) {
	rows := []EventRow{}

	err := r.db.SelectContext(ctx, &rows, `SELECT id, event FROM events WHERE ($2 = '' OR event->>'category_id' = $2) ORDER BY event->>'city_id' = $1 DESC, event->>'date' ASC LIMIT $4 OFFSET $3`,
		cityID, category, startKey, count)
	if err != nil {
		return nil, err
	}

	eventSlice := make([]models.Event, len(rows))

	for i := range rows {
		eventSlice[i] = *rows[i].Event
		eventSlice[i].ID = rows[i].ID
	}

	return eventSlice, nil
}
