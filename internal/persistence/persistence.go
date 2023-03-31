package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	models "github.com/Iknite-space/cliqets-api/internal/models/event"
	"github.com/rs/zerolog"

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
	GetEventByID(ctx context.Context, id string) (*models.Event, error)
	GetUpComingEventsCities(ctx context.Context) ([]models.City, error)
	CreateUser(ctx context.Context, newUser models.User) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	GetPurchasedTicketBYID(ctx context.Context, purchasedID string) (*models.PurchasedTicket, error)
	GetPurchasedTickets(ctx context.Context, userID string, eventID string) ([]models.PurchasedTicket, error)
	UpdateUserInfo(ctx context.Context, user models.User, userID string) (*models.User, error)
	UpdateEvent(ctx context.Context, event models.Event, eventID string) (*models.Event, error)
	CreateOrder(ctx context.Context, order models.Order) (*models.Order, error)
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	CreatePurchasedTicket(ctx context.Context, purchasedOrder models.PurchasedTicket) (*models.PurchasedTicket, error)
	GetAllCategories(ctx context.Context) ([]string, error)
	GetGuestList(ctx context.Context, eventID string) ([]models.GuestListEntry, error)
	CreateEvent(ctx context.Context, event models.Event) (*models.Event, error)
	UpdateOrderStatus(ctx context.Context, status string, orderID string) (*models.Order, error)
}

// logger.
var logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

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

// GetEventByID implements Repository.
func (r *RepositoryPg) GetEventByID(ctx context.Context, id string) (*models.Event, error) {
	row := EventRow{}

	err := r.db.GetContext(ctx, &row, "SELECT id, event FROM events WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	row.Event.ID = row.ID

	return row.Event, nil
}

func (r *RepositoryPg) GetUpComingEventsCities(ctx context.Context) ([]models.City, error) {
	currentCity := []models.City{}
	err := r.db.SelectContext(ctx, &currentCity, "SELECT * FROM cities WHERE city_id in (SELECT event->>'city_id' AS cities from events)")
	//nolintlint:wsl
	if err != nil {
		return nil, err
	}

	return currentCity, nil
}

func (r *RepositoryPg) CreateUser(ctx context.Context, newUser models.User) (*models.User, error) {
	createdUser := models.User{}

	err := r.db.GetContext(ctx, &createdUser, "INSERT INTO users (user_id, first_name,last_name, phone_number, email, profile_image,current_city, country, gender) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING user_id, first_name, last_name, phone_number, email,profile_image, current_city, country, gender",
		newUser.UserID, newUser.FirtName, newUser.LastName, newUser.PhoneNumber, newUser.Email, newUser.ProfileImage, newUser.CityID, newUser.Country, newUser.Gender)
	if err != nil {
		return nil, err
	}

	return &createdUser, nil
}

func (r *RepositoryPg) GetUser(ctx context.Context, userID string) (*models.User, error) {
	user := models.User{}

	err := r.db.GetContext(ctx, &user, "SELECT user_id, first_name,last_name, phone_number, email, profile_image,current_city, country, gender FROM users WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// get all purchased tickets for an event.
func (r *RepositoryPg) GetPurchasedTickets(ctx context.Context, userID string, eventID string) ([]models.PurchasedTicket, error) {
	purchasedTickets := []models.PurchasedTicket{}

	err := r.db.SelectContext(ctx, &purchasedTickets, `SELECT events.id as "event_id", event->>'title' as "title", event->>'cover_photo' as "cover_image", event->>'venue' as "venue",
	event->>'organizer' as "organizer_name",event->>'Date' as "event_date",
	purchased_tickets.purcchase_ticket_id,
	purchased_tickets.ticket_no,
	purchased_tickets.hall_no,
	purchased_tickets.order_no,
	purchased_tickets.seat_no,
	purchased_tickets.ticket_type,
	purchased_tickets.user_id
	FROM events INNER JOIN purchased_tickets 
	ON events.id=purchased_tickets.event_id
	WHERE  user_id=$1 AND ($2='' OR event_id=$2::uuid)
	ORDER BY event->>'Date' ASC`, userID, eventID)
	if err != nil {
		errMsg := "failed to get purchased ticket "
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :-> %v\n", errMsg, err)

		return nil, fmt.Errorf("%s :-> %w", errMsg, err)
	}

	return purchasedTickets, nil
}

// update user info.
func (r *RepositoryPg) UpdateUserInfo(ctx context.Context, user models.User, userID string) (*models.User, error) {
	info := models.User{}

	err := r.db.GetContext(ctx, &info, "UPDATE users SET first_name=$1, last_name=$2, profile_image=$3,current_city=$4, country=$5, email=$6, gender=$7  WHERE user_id=$8 RETURNING user_id, first_name, last_name, phone_number, gender, email, profile_image, current_city, country", user.FirtName, user.LastName, user.ProfileImage, user.CityID, user.Country, user.Email, user.Gender, userID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// create new order.
func (r *RepositoryPg) CreateOrder(ctx context.Context, order models.Order) (*models.Order, error) {
	newOrder := models.Order{}

	err := r.db.GetContext(ctx, &newOrder, "INSERT INTO orders (user_id,event_id,user_name,payment_provider,number_of_tickets,amount,order_date,phone_number,purchase_status,ticket) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)RETURNING order_id,user_id,event_id,payment_provider,number_of_tickets,amount,order_date,phone_number,purchase_status,ticket", order.UserID, order.EventID, order.UserName, order.PaymentProvider, order.NumberOfTickets, order.Amount, order.OrderDate, order.PhoneNumber, order.PurchaseStatus, order.Ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to persist create order: %w", err)
	}

	order.OrderID = newOrder.OrderID

	return &order, nil
}

// Get the order status for a particular order from a particular user

func (r *RepositoryPg) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	orderStatus := models.Order{}

	err := r.db.GetContext(ctx, &orderStatus, "SELECT * FROM orders  WHERE order_id = $1", orderID)
	if err != nil {
		return nil, err
	}

	return &orderStatus, nil
}

func (r *RepositoryPg) CreatePurchasedTicket(ctx context.Context, pOrder models.PurchasedTicket) (*models.PurchasedTicket, error) {
	newPurchasedOrder := &models.PurchasedTicket{}

	err := r.db.GetContext(ctx, newPurchasedOrder, "INSERT INTO purchased_tickets (ticket_no,hall_no,order_no,seat_no,ticket_type,user_id,event_id) VALUES ($1,$2,$3,$4,$5,$6,$7)RETURNING purcchase_ticket_id, event_id, ticket_type", pOrder.TicketNum, pOrder.HallNum, pOrder.OrderNo, pOrder.SeatNum, pOrder.TicketType, pOrder.UserID, pOrder.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to persist purchased tickets: %w", err)
	}

	return newPurchasedOrder, nil
}

func (r *RepositoryPg) UpdateOrderStatus(ctx context.Context, status, orderID string) (*models.Order, error) {
	order := &models.Order{}

	err := r.db.GetContext(ctx, order, "UPDATE orders SET purchase_status=$1 WHERE order_id=$2 RETURNING order_id,user_id,event_id,payment_provider,number_of_tickets,amount,order_date,phone_number,purchase_status,ticket", status, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create purchased tickets: %w", err)
	}

	return order, nil
}

func (r *RepositoryPg) GetAllCategories(ctx context.Context) ([]string, error) {
	categories := []string{}
	err := r.db.SelectContext(ctx, &categories, "SELECT DISTINCT event->>'category_id' FROM events")
	// nolintlint:wsl
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *RepositoryPg) CreateEvent(ctx context.Context, event models.Event) (*models.Event, error) {
	row := EventRow{}
	err := r.db.GetContext(ctx, &row, "iNSERT INTO	events (event) VALUES($1) RETURNING id, event", event)
	//nolintlint:wsl
	if err != nil {
		return nil, fmt.Errorf("couldn't create event %w", err)
	}

	row.Event.ID = row.ID

	return row.Event, nil
}

func (r *RepositoryPg) GetPurchasedTicketBYID(ctx context.Context, purchasedID string) (*models.PurchasedTicket, error) {
	purchasedTicket := models.PurchasedTicket{}
	err := r.db.GetContext(ctx, &purchasedTicket, "SELECT * FROM purchased_tickets WHERE purcchase_ticket_id=$1", purchasedID)
	// nolintlint:wsl
	if err != nil {
		return nil, fmt.Errorf("failed to get purchased ticket by id %w", err)
	}

	return &purchasedTicket, nil
}

func (r *RepositoryPg) GetGuestList(ctx context.Context, eventID string) ([]models.GuestListEntry, error) {
	guestListTicket := []models.GuestListEntry{}

	err := r.db.SelectContext(ctx, &guestListTicket, `SELECT first_name, last_name, phone_number,
	purchased_tickets.purcchase_ticket_id,
	purchased_tickets.ticket_no,
	purchased_tickets.hall_no,
	purchased_tickets.order_no,
	purchased_tickets.seat_no,
	purchased_tickets.ticket_type,
	purchased_tickets.user_id
	FROM users INNER JOIN purchased_tickets 
	ON users.user_id=purchased_tickets.user_id
	WHERE  purchased_tickets.event_id=$1`, eventID)
	if err != nil {
		errMsg := "failed to get purchased ticket "
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :-> %v\n", errMsg, err)

		return nil, fmt.Errorf("%s :-> %w", errMsg, err)
	}

	return guestListTicket, nil
}

func (r *RepositoryPg) UpdateEvent(ctx context.Context, patchEvent models.Event, eventID string) (*models.Event, error) {
	row := EventRow{}

	err := r.db.GetContext(ctx, &row, "UPDATE events SET event =$1 WHERE id=$2 RETURNING id, event", patchEvent, eventID)
	if err != nil {
		return nil, err
	}

	row.Event.ID = row.ID

	return row.Event, nil
}
