package trips

import "time"

const (
	//ParticipantOwnerRole defines a participant as a trip owner
	ParticipantOwnerRole = "owner"
	//ParticipantAdminRole defines a participant as a trip admin
	ParticipantAdminRole = "admin"
	//ParticipantEditorRole defines a participant as a trip editor
	ParticipantEditorRole = "editor"
	//ParticipantViewerRole defines a participant as a trip viewer
	ParticipantViewerRole = "viewer"
)

// Participant represents a user that is participating in the trip
type Participant struct {
	ID          string    `json:"id" db:"id" lock:"true"`
	TripID      string    `json:"trip_id" db:"trip_id" lock:"true"`
	UserID      string    `json:"user_id" db:"user_id" lock:"true"`
	Role        string    `json:"role" db:"role"`
	CreatedBy   string    `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	UpdatedDate time.Time `json:"updated_date" db:"updated_date"`
}
