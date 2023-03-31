package events

import (
	"context"
	"fmt"
	"os"
	"time"

	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/event"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/persistence"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/services/payments"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source ./events.go -destination mocks/events.mock.go -package mocks

// Service provide fuctionality for managing events and tickets.
//
//nolint:interfacebloat
type Service interface {
	GetAllEvents(ctx context.Context, cityID string, categoryID string, startKey uint, count uint) ([]models.Event, error)
}

type ServiceImpl struct {
	repo       persistence.Repository
	pgGateway  payments.PaymentService
	webHookKey string
}

var (
	ErrUnepectedSigningAlg = fmt.Errorf("unexpected signing algorithm")
	// logs in json format.
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

//nolint:exhaustivestruct
var _ Service = &ServiceImpl{}

func NewService(repo persistence.Repository, pgGateway payments.PaymentService, webHookAppKey string) (*ServiceImpl, error) {
	return &ServiceImpl{
		repo:       repo,
		pgGateway:  pgGateway,
		webHookKey: webHookAppKey,
	}, nil
}

// GetAllEvents implements Service.
func (s *ServiceImpl) GetAllEvents(ctx context.Context, cityID string, categoryID string, startKey uint, count uint) ([]models.Event, error) {
	filteredEvents := []models.Event{}

	events, err := s.repo.GetEvents(ctx, cityID, categoryID, startKey, count)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("cityID", cityID).Str("categoryID", categoryID).Str("startKey", fmt.Sprint(startKey)).Msg("Call to repo.GetEvents failed")

		return nil, err
	}

	// Reminder: write test for filtering events and track events id for archiving functionality
	for _, event := range events {
		// grace period of 24 hours
		const hours = 24
		graceTime := time.Now().Add(-time.Hour * hours)

		if graceTime.String() <= event.Date {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}
