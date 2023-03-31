package events

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	paymentModels "github.com/Iknite-space/cliqets-api/internal/models/payments"
	"github.com/Iknite-space/cliqets-api/internal/persistence"
	"github.com/Iknite-space/cliqets-api/internal/services/payments"
	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source ./events.go -destination mocks/events.mock.go -package mocks

// Service provide fuctionality for managing events and tickets.
//
//nolint:interfacebloat
type Service interface {
	GetAllEvents(ctx context.Context, cityID string, categoryID string, startKey uint, count uint) ([]models.Event, error)
	GetEventByID(ctx context.Context, id string) (*models.Event, error)
	GetCurrentCity(ctx context.Context, long float64, lat float64) (*models.City, error)
	CreateUser(ctx context.Context, newUser models.User) (*models.User, error)
	GetUser(ctx context.Context, uiud string) (*models.User, error)
	GetPurchasedTickets(ctx context.Context, userID string, eventID string) ([]models.PurchasedTicket, error)
	GetGuestList(ctx context.Context, eventID string) ([]models.GuestListEntry, error)
	GetPurchasedTicketBYID(ctx context.Context, purchasedID string) (*models.PurchasedTicket, error)
	GetBookedEvents(ctx context.Context, userID string) (*models.BookedEvents, error)
	UpdateEvent(ctx context.Context, event models.Event, eventID string) (*models.Event, error)
	UpdateUserInfo(ctx context.Context, user models.User, userID string) (*models.User, error)
	CreateOrder(ctx context.Context, order models.Order) (*models.Order, error)
	TransStatus(ctx context.Context, status string, exRef string, amount string, currency string, code string, operatorRef string, signature string) (transOrder *models.Order, err error)
	GetOrderByID(ctx context.Context, orderID string, userID string) (*models.Order, error)
	GetAllCategories(ctx context.Context) ([]string, error)
	CreateEvent(ctx context.Context, event models.Event) (*models.Event, error)
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

// GetEventByID implements Service.
func (s *ServiceImpl) GetEventByID(ctx context.Context, id string) (*models.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("eventID", id).Msg("repo.GetEventByID service failed")

		return nil, err
	}

	return event, nil
}

func (s *ServiceImpl) GetCurrentCity(ctx context.Context, long float64, lat float64) (*models.City, error) {
	cities, err := s.repo.GetUpComingEventsCities(ctx)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("Longitude", fmt.Sprint(long)).Str("Latitude", fmt.Sprint(lat)).Msgf("repo.Getupcoming events failed")
		return nil, err
	}

	if len(cities) < 1 {
		//nolint:goerr113
		return nil, fmt.Errorf("no nearby city found")
	}

	// calc min distance
	smallestCityIndex := 0
	minDistance := distance(cities[0].Latitude, cities[0].Longitute, lat, long, "K")

	// calc distance between cities with upconing events and user lng/lat in Km
	for index := 0; index < len(cities); index++ {
		citiesDistance := distance(cities[0].Latitude, cities[0].Longitute, lat, long, "K")
		if citiesDistance < minDistance {
			minDistance = citiesDistance
			smallestCityIndex = index
		}
	}

	return &cities[smallestCityIndex], nil
}

// implement create user feature.
func (s *ServiceImpl) CreateUser(ctx context.Context, newUser models.User) (*models.User, error) {
	user, err := s.repo.CreateUser(ctx, newUser)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("User", fmt.Sprint(newUser)).Msg("repo.CreateUser service failed")

		return nil, err
	}

	return user, nil
}

// GetUser on sign in.
func (s *ServiceImpl) GetUser(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("UserID", userID).Msg("repo.GetUser service failed")

		return nil, err
	}

	return user, nil
}

// get all purchased tickets for an event.
func (s *ServiceImpl) GetPurchasedTickets(ctx context.Context, userID string, eventID string) ([]models.PurchasedTicket, error) {
	purchedTickets, err := s.repo.GetPurchasedTickets(ctx, userID, eventID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("UserID", userID).Str("EventID", eventID).Msg("repo.GetPurchasedTickets service failed")

		return nil, err
	}

	return purchedTickets, nil
}

// aggregate purchased tickets.
func (s *ServiceImpl) GetBookedEvents(ctx context.Context, userID string) (*models.BookedEvents, error) {
	groupedPassEvents := []models.EventTicketsSummary{}
	groupedFutureEvents := []models.EventTicketsSummary{}

	tickets, err := s.repo.GetPurchasedTickets(ctx, userID, "")
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("UserID", userID).Msg("repo.GetBookedEvents service failed")

		return nil, err
	}

	allEventsSummary := map[string]*models.EventTicketsSummary{}

	// implement aggregation logic service
	for _, ticket := range tickets {
		if _, exist := allEventsSummary[ticket.EventID]; exist {
			allEventsSummary[ticket.EventID].NumTickets++
			continue
		}

		allEventsSummary[ticket.EventID] = &models.EventTicketsSummary{
			Title:      ticket.Title,
			CoverImg:   ticket.CoverImg,
			Venue:      ticket.Venue,
			NumTickets: 1,
			Date:       ticket.EventDate,
			EventID:    ticket.EventID,
		}
	}

	// sort tickets by future and pass event.
	for _, ticket := range tickets {
		// check if already seen ticket ?
		eventSummary, exists := allEventsSummary[ticket.EventID]
		if !exists {
			continue
		}

		delete(allEventsSummary, ticket.EventID)

		if eventSummary.Date <= time.Now().String() {
			groupedPassEvents = append([]models.EventTicketsSummary{*eventSummary}, groupedPassEvents...)
		} else {
			groupedFutureEvents = append(groupedFutureEvents, *eventSummary)
		}
	}

	return &models.BookedEvents{FutureEvents: groupedFutureEvents, PastEvents: groupedPassEvents}, nil
}

func (s *ServiceImpl) UpdateUserInfo(ctx context.Context, user models.User, userID string) (*models.User, error) {
	upUser, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("User", fmt.Sprint(user)).Str("userID", userID).Msg("repo.UpdateUserInfo service failed")

		return nil, err
	}

	if user.FirtName != "" {
		upUser.FirtName = user.FirtName
	}

	if user.LastName != "" {
		upUser.LastName = user.LastName
	}

	// if user.DOB != "" {
	// 	upUser.DOB = user.DOB
	// }

	if user.Gender != "" {
		upUser.Gender = user.Gender
	}

	if user.Email != "" {
		upUser.Email = user.Email
	}

	if user.Country != "" {
		upUser.Country = user.Country
	}

	if user.CityID != "" {
		//nolint:godox
		//Todo: change type in Model and acccess as user.City.name
		upUser.CityID = user.CityID
	}

	updateInfo, err := s.repo.UpdateUserInfo(ctx, *upUser, userID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("User", fmt.Sprint(updateInfo)).Str("UserID", userID).Msg("repo.UpdateUserInfo service failed")

		return nil, err
	}

	return updateInfo, nil
}

func (s *ServiceImpl) CreateOrder(ctx context.Context, order models.Order) (*models.Order, error) {
	events, err := s.repo.GetEventByID(ctx, order.EventID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("Event", events.ID).Msg("call to repo.GetEventByID  failed on createOrder service")

		return nil, fmt.Errorf("%w", err)
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Str("Event", events.ID).Msg("Successfully fetched Event")

	order.Amount = 0

	for index, ticket := range order.Ticket {
		// lookup ticket types and return price
		sum, price, err := getTicketPrice(events.Ticket, ticket.TicketType, ticket.Quantity)
		if err != nil {
			logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("Sum", fmt.Sprint(sum)).Str("Price", fmt.Sprint(price)).Msg("getTicketPrice failed to calculate price and sum in create order service")

			return nil, fmt.Errorf("%w", err)
		}

		logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully calculated ticket price")

		order.PurchaseStatus = "PENDING"
		order.Amount += sum
		ticket.Price = price
		order.Ticket[index] = ticket
	}

	newOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("newOrder", fmt.Sprint(newOrder)).Msg("call to repo.CreateOrder service")

		return nil, err
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully created order")
	// build request struct
	initiatePymnts := paymentModels.RequestBody{
		Amount:      fmt.Sprintf("%d", newOrder.Amount),
		From:        newOrder.PhoneNumber,
		Description: events.Title,
		ExternalRef: newOrder.OrderID,
	}
	// initiate pyments
	err = s.pgGateway.InitiatePayments(ctx, initiatePymnts)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("initiatePymnts", fmt.Sprint(initiatePymnts)).Msg("call to s.pgGateway.InitiatePayments service failed")

		return nil, err
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully made payment request to Campay")

	// pymnts response
	return newOrder, err
}

//nolint:funlen
func (s *ServiceImpl) TransStatus(ctx context.Context, status string, exRef string, amount string, currency string, code string, opRef string, signature string) (*models.Order, error) {
	token, err := jwt.Parse(signature, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			alg := token.Header["alg"]
			return nil, fmt.Errorf("unexpected signing algorithm: alg=%v %w", alg, ErrUnepectedSigningAlg)
		}
		return []byte(s.webHookKey), nil
	})

	// validate jwt token
	if !token.Valid {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msg("Token is not valid")

		return nil, fmt.Errorf("%w", err)
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully confirmed tocken")

	logger.Debug().Msg("Successfully confirmed tocken")

	if status != "SUCCESSFUL" {
		// update order status
		orderStatus, err := s.repo.UpdateOrderStatus(ctx, status, exRef)
		if err != nil {
			logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msg("call to repo.UpdateOrderStatus failed in TransStatus service")

			return nil, fmt.Errorf("%w", err)
		}

		logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Payment Successful")

		return orderStatus, nil
	}

	order, err := s.repo.GetOrderByID(ctx, exRef)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msg("call to repo.GetOrderByID failed in TransStatus service")

		return nil, fmt.Errorf("%w", err)
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("successfully fetched order")

	//  check amt from web hook response
	//nolint
	resAmount, err := strconv.ParseFloat(amount, 32)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msgf("couldn't parse amount from web hook:: %vs", err)

		return nil, err
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully parsed amount from web hook")

	if float64(order.Amount) != resAmount {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("Order Amount", fmt.Sprint(order.Amount)).Str("ResAmount", fmt.Sprint(resAmount)).Msg("Amount from web hook not same with Order Amount")

		return nil, fmt.Errorf("%w", err)
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Amount from web hook matched amount in the order")

	for _, pTickets := range order.Ticket {
		for i := 0; i < int(pTickets.Quantity); i++ {
			// create purchased tickets
			//nolint:exhaustruct
			pOrderSummary := models.PurchasedTicket{
				EventID:    order.EventID,
				TicketType: pTickets.TicketType,
				UserID:     order.UserID,
				OrderID:    order.OrderID,
			}

			_, err := s.repo.CreatePurchasedTicket(ctx, pOrderSummary)
			if err != nil {
				logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msgf("call to repo.CreatePurchasedTicket failed in TransStatus service")

				return nil, fmt.Errorf("%w", err)
			}

			logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully created ticket")
		}
	}

	// update order status
	orderStatus, err := s.repo.UpdateOrderStatus(ctx, status, exRef)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msgf("call to repo.UpdateOrderStatus failed in TransStatus service")

		return nil, fmt.Errorf("%w", err)
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully updated order status")

	return orderStatus, nil
}

func (s *ServiceImpl) GetOrderByID(ctx context.Context, orderID string, userID string) (*models.Order, error) {
	orderStatus, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("OrderID", orderID).Str("UserID", userID).Msgf("call to repo.GetOrderByID service failed ")

		return nil, err
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Successfully fetched order")

	if orderStatus.UserID != userID {
		//nolint:goerr113
		return nil, fmt.Errorf("order not found")
	}

	logger.Debug().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msg("Order exists in order table")

	return orderStatus, nil
}

func (s *ServiceImpl) GetAllCategories(ctx context.Context) ([]string, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Msg("call to repo.GetAllCategories service failed")

		return nil, err
	}

	return categories, nil
}

func (s *ServiceImpl) CreateEvent(ctx context.Context, event models.Event) (*models.Event, error) {
	events, err := s.repo.CreateEvent(ctx, event)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Err(err).Str("Event", fmt.Sprint(event)).Msg("repo.CreateEvent service failed")

		return nil, err
	}

	return events, nil
}

func (s *ServiceImpl) GetPurchasedTicketBYID(ctx context.Context, purchasedID string) (*models.PurchasedTicket, error) {
	purchasedTicket, err := s.repo.GetPurchasedTicketBYID(ctx, purchasedID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Str("Purchased ID", purchasedID).Msg("call to repo.GetPurchasedTicketBYID service failed")
		return nil, err
	}

	return purchasedTicket, nil
}

func (s *ServiceImpl) GetGuestList(ctx context.Context, eventID string) ([]models.GuestListEntry, error) {
	guestTicketsEntries, err := s.repo.GetGuestList(ctx, eventID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Str("Event ID", eventID).Msg("call to repo.GetGuestList service failed")
		return nil, err
	}

	return guestTicketsEntries, nil
}

//nolint:funlen
func (s *ServiceImpl) UpdateEvent(ctx context.Context, patchEvent models.Event, eventID string) (*models.Event, error) {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Str("Event ID", eventID).Msg("call to repo.GetEventByID service failed")
		return nil, err
	}
	// Reminder :- improve implementation.
	if patchEvent.Title != "" {
		event.Title = patchEvent.Title
	}

	if patchEvent.City != "" {
		event.City = patchEvent.City
	}

	if patchEvent.Venue != "" {
		event.Venue = patchEvent.Venue
	}

	if patchEvent.CityID != "" {
		event.CityID = patchEvent.CityID
	}

	if patchEvent.CoverPhoto != "" {
		event.CoverPhoto = patchEvent.CoverPhoto
	}

	if patchEvent.Description != "" {
		event.Description = patchEvent.Description
	}

	if patchEvent.Organiser != "" {
		event.Organiser = patchEvent.Organiser
	}

	if patchEvent.CategoryID != "" {
		event.CategoryID = patchEvent.CategoryID
	}

	if patchEvent.Time != "" {
		event.Time = patchEvent.Time
	}

	if patchEvent.Date != "" {
		event.Date = patchEvent.Date
	}

	if patchEvent.HeadLine != "" {
		event.HeadLine = patchEvent.HeadLine
	}

	if patchEvent.RefundPolicy != "" {
		event.RefundPolicy = patchEvent.RefundPolicy
	}

	if patchEvent.Ticket != nil {
		event.Ticket = patchEvent.Ticket
	}

	updatedEvent, err := s.repo.UpdateEvent(ctx, *event, eventID)
	if err != nil {
		return nil, err
	}

	return updatedEvent, nil
}
