package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Cars struct {
	ID                string `json:"id"`
	SellerID          string `json:"seller_id"`
	CarName           string `json:"car_name"`
	DatePosted        string `json:"date_posted"`
	BidingPrice       string `json:"biding_price"`
	BidExpirationTime string `json:"bid_expiration_time"`
	CityID            string `json:"city_id"`
	EngineType        string `json:"engine_type"`
	CarModel          string `json:"car_model"`
	Mileage           string `json:"mileage"`
	FuelType          string `json:"fuel_type"`
	CarphotoUrl       string `json:"photo_url"`
	Category          string `json:"category"`
	Description       string `json:"description"`
}
type Sellers struct {
	User_id  string `json:"id" db:"user_id"`
	UserName string `json:"user_name" db:"user_name"`
	Email    string `json:"user_email" db:"user_email"`
}
type Bids struct {
	BidID    string `json:"bid_id" db:"bid_id"`
	CarID    string `json:"car_id" db:"car_id"`
	Amount   string `json:"bid_amount" db:"bid_amount"`
	Email    string `json:"email" db:"email"`
	UserName string `json:"user_name" db:"user_name"`
}

func (e Cars) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *Cars) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		//nolint:goerr113
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &e)
}
