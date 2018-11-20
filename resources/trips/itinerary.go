package trips

import (
	"time"

	"github.com/feedmytrip/api/resources/shared"
)

// Itinerary represents a way to group trip events
type Itinerary struct {
	ID          string             `json:"id" db:"id" lock:"true"`
	TripID      string             `json:"trip_id" db:"trip_id" lock:"true"`
	Title       shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = trip_itinerary.id and title.field = 'title'" embedded:"true" persist:"true"`
	OwnerID     string             `json:"owner_id" db:"owner_id"`
	StartDate   time.Time          `json:"start_date" db:"start_date"`
	EndDate     time.Time          `json:"end_date" db:"end_date"`
	CreatedBy   string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy   string             `json:"updated_by" db:"updated_by"`
	UpdatedDate time.Time          `json:"updated_date" db:"updated_date"`
}
