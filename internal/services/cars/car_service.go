package cars

import (
	"context"
	"fmt"

	models "github.com/namkatcedrickjumtock/sigma-auto-api/internal/models/cars"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/persistence"
	"github.com/namkatcedrickjumtock/sigma-auto-api/internal/services/payments"
)

//go:generate mockgen -source ./events.go -destination mocks/events.mock.go -package mocks

// Service provide fuctionality for managing events and tickets.
//
//nolint:interfacebloat
type Service interface {
	GetAllCars(ctx context.Context, cityID string, category string, startKey uint, count uint) ([]models.Cars, error)
	// UpdateCar(ctx context.Context, updatePayLoad models.Cars, carID string) (*models.Cars, error)
	RegisterCar(ctx context.Context, carPayload models.Cars) (*models.Cars, error)
	GetCarsByID(ctx context.Context, carID string) (*models.Cars, error)
	PlaceBid(ctx context.Context, bid models.Bids) (*models.Bids, error)
	GetBidByID(ctx context.Context, bidID string) (*models.Bids, error)
	GetUserByID(ctx context.Context, userID string) (*models.Users, error)
	CreateUser(ctx context.Context, user models.Users) (*models.Users, error)
}

type ServiceImpl struct {
	repo       persistence.Repository
	pgGateway  payments.PaymentService
	webHookKey string
}

var (
	ErrUnepectedSigningAlg = fmt.Errorf("unexpected signing algorithm")
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
func (s *ServiceImpl) GetAllCars(ctx context.Context, cityID string, categoryID string, startKey uint, count uint) ([]models.Cars, error) {

	cars, err := s.repo.GetAllCars(ctx, cityID, categoryID, startKey, count)
	if err != nil {

		return nil, err
	}

	return cars, nil
}

// func (s *ServiceImpl) UpdateCar(ctx context.Context, updatePayLoad models.Cars, carID string) (*models.Cars, error) {
// }
func (s *ServiceImpl) RegisterCar(ctx context.Context, carPayload models.Cars) (*models.Cars, error) {
	newRegisteredCar, err := s.repo.RegisterCar(ctx, carPayload)
	if err != nil {
		return nil, err
	}
	return newRegisteredCar, nil
}

func (s *ServiceImpl) GetCarsByID(ctx context.Context, carID string) (*models.Cars, error) {
	car, err := s.repo.GetCarsByID(ctx, carID)
	if err != nil {
		return nil, err
	}
	return car, nil
}

func (s *ServiceImpl) PlaceBid(ctx context.Context, bid models.Bids) (*models.Bids, error) {
	bids, err := s.repo.PlaceBid(ctx, bid)
	if err != nil {
		return nil, err
	}
	return bids, nil
}

func (s *ServiceImpl) GetBidByID(ctx context.Context, bidID string) (*models.Bids, error) {
	bid, err := s.repo.GetBidByID(ctx, bidID)
	if err != nil {
		return nil, err
	}
	return bid, nil
}

// CreateUser(ctx context.Context, user models.Sellers) (*models.Sellers, error)
func (s *ServiceImpl) GetUserByID(ctx context.Context, userID string) (*models.Users, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *ServiceImpl) CreateUser(ctx context.Context, user models.Users) (*models.Users, error) {
	newUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

