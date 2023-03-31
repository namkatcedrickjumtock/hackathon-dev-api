package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Event struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	City         string   `json:"city"`
	HeadLine     string   `json:"headline"`
	CityID       string   `json:"city_id"`
	Date         string   `json:"Date"`
	Time         string   `json:"time"`
	Venue        string   `json:"venue"`
	CoverPhoto   string   `json:"cover_photo"`
	Ticket       []Ticket `json:"tickets"`
	Description  string   `json:"description"`
	Organiser    string   `json:"organizer"`
	RefundPolicy string   `json:"refund_policy"`
	CategoryID   string   `json:"category_id"`
}

type Ticket struct {
	TicketID    string `json:"id"`
	Name        string `json:"name"`
	TicketImage string `json:"ticket_image"`
	Price       int64  `json:"price"`
	Description string `json:"description"`
	Status      bool   `json:"status"`
	TicketLimit int64  `json:"ticket_limit"`
}

type City struct {
	CityID    string  `json:"city_id" db:"city_id"`
	Name      string  `json:"name" db:"city_name"`
	Longitute float64 `json:"longitude" db:"longitude"`
	Latitude  float64 `json:"latitude" db:"latitude"`
}

type PurchasedTicket struct {
	PurchaseTicket string `json:"purcchase_ticket_id" db:"purcchase_ticket_id"`
	EventID        string `json:"event_id" db:"event_id"`
	OrderID        string `json:"order_id" db:"order_id"`
	Title          string `json:"title" db:"title"`
	OrganizerName  string `json:"organizer_name" db:"organizer_name"`
	TicketNum      string `json:"ticket_no" db:"ticket_no"`
	HallNum        int64  `json:"hall_no" db:"hall_no"`
	OrderNo        string `json:"order_no" db:"order_no"`
	SeatNum        int64  `json:"seat_no" db:"seat_no"`
	TicketType     string `json:"ticket_type" db:"ticket_type"`
	UserID         string `json:"user_id" db:"user_id"`
	CoverImg       string `json:"cover_image" db:"cover_image"`
	Venue          string `json:"venue" db:"venue"`
	EventDate      string `json:"event_date" db:"event_date"`
}

type GuestListEntry struct {
	PurchaseTicket string `json:"purcchase_ticket_id" db:"purcchase_ticket_id"`
	EventID        string `json:"event_id" db:"event_id"`
	OrderID        string `json:"order_id" db:"order_id"`
	Title          string `json:"title" db:"title"`
	OrganizerName  string `json:"organizer_name" db:"organizer_name"`
	TicketNum      string `json:"ticket_no" db:"ticket_no"`
	HallNum        int64  `json:"hall_no" db:"hall_no"`
	OrderNo        string `json:"order_no" db:"order_no"`
	SeatNum        int64  `json:"seat_no" db:"seat_no"`
	TicketType     string `json:"ticket_type" db:"ticket_type"`
	UserID         string `json:"user_id" db:"user_id"`
	CoverImg       string `json:"cover_image" db:"cover_image"`
	Venue          string `json:"venue" db:"venue"`
	EventDate      string `json:"event_date" db:"event_date"`
	FirtName       string `json:"first_name" db:"first_name"`
	LastName       string `json:"last_name" db:"last_name"`
	PhoneNumber    string `json:"phone_number" db:"phone_number"`
}
type BookedEvents struct {
	FutureEvents []EventTicketsSummary `json:"future_events"`
	PastEvents   []EventTicketsSummary `json:"past_events"`
}

type EventTicketsSummary struct {
	Title      string `json:"title" db:"title"`
	CoverImg   string `json:"cover_image" db:"cover_image"`
	Venue      string `json:"venue" db:"venue"`
	Date       string `json:"date" db:"date"`
	NumTickets int64  `json:"number_of_tickets" db:"number_of_tickets"`
	EventID    string `json:"event_id" db:"event_id"`
}

type User struct {
	UserID      string `json:"user_id" db:"user_id"`
	FirtName    string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
	Email       string `json:"email" db:"email"`
	Gender      string `json:"gender" db:"gender"`
	// REMINDER :- fix {"error":"pq: invalid input syntax for type date: \"\""}.
	//  can't accept a string when creating user.
	// DOB          string `json:"date_of_birth" db:"date_of_birth"`.
	CityID       string `json:"current_city" db:"current_city"`
	ProfileImage string `json:"profile_image" db:"profile_image"`
	Country      string `json:"country" db:"country"`
}

type Order struct {
	OrderID         string           `json:"order_id" db:"order_id"`
	UserID          string           `json:"user_id" db:"user_id"`
	EventID         string           `json:"event_id" db:"event_id"`
	UserName        string           `json:"user_name" db:"user_name"`
	PaymentProvider string           `json:"payment_provider" db:"payment_provider"`
	NumberOfTickets string           `json:"number_of_tickets" db:"number_of_tickets"`
	Amount          int64            `json:"amount" db:"amount"`
	OrderDate       string           `json:"order_date" db:"order_date"`
	PhoneNumber     string           `json:"phone_number" db:"phone_number"`
	PurchaseStatus  string           `json:"purchase_status" db:"purchase_status"`
	Ticket          OrderTicketTypes `json:"ticket" db:"ticket"`
}

type OrderTicketTypes []OrderTicketType

type OrderTicketType struct {
	TicketType string `json:"ticket_type" db:"ticket_type"`
	Quantity   int64  `json:"quantity" db:"quantity"`
	Price      int64  `json:"price" db:"price"`
}

func (o OrderTicketTypes) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OrderTicketTypes) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		//nolint:goerr113
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &o)
}

func (e Event) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *Event) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		//nolint:goerr113
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &e)
}
