package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Cars struct {
	ID          string `json:"id"`
	SellerID    string `json:"seller_id"`
	CarName     string `json:"car_name"`
	DatePosted  string `json:"date_posted"`
	Time        string `json:"time"`
	CityID      string `json:"city_id"`
	CarphotoUrl string `json:"photo_url"`
	Category    string `json:"category"`
	Description string `json:"description"`
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
