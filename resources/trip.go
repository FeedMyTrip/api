package resources

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

// Trip represents a user trip
type Trip struct {
	TripID      string    `json:"tripId" validate:"required"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	CreatedDate time.Time `json:"createdDate" validate:"required"`
}

// NewTrip returns a new Trip pointer with a unique id
func NewTrip(body string) (*Trip, error) {
	t := &Trip{}
	err := json.Unmarshal([]byte(body), t)
	if err != nil {
		return nil, err
	}
	t.TripID = uuid.New().String()
	t.CreatedDate = time.Now()

	validate := validator.New()
	err = validate.Struct(t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
