package resources

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

// Participant represents a user that is participating in the trip
type Participant struct {
	ParticipantID string    `json:"participantId" validate:"required"`
	UserID        string    `json:"userId" validate:"required"`
	UserRole      string    `json:"userRole" validate:"required"`
	CreatedBy     string    `json:"createdBy"`
	CreatedDate   time.Time `json:"createdDate"`
	UpdatedBy     string    `json:"updatedBy"`
	UpdatedDate   time.Time `json:"updatedDate"`
}

// NewParticipant returns a new participant pointer with a unique id
func NewParticipant(body string) (*Participant, error) {
	p := &Participant{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		return nil, err
	}
	p.ParticipantID = uuid.New().String()
	//TODO replace 000001 by the userID that execute the action from Cognito
	p.CreatedBy = "000001"
	p.CreatedDate = time.Now()
	p.UpdatedBy = "000001"
	p.UpdatedDate = time.Now()

	validate := validator.New()
	err = validate.Struct(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func newParticipantOwner(ownerID string) *Participant {
	p := &Participant{}
	p.ParticipantID = uuid.New().String()
	p.UserID = ownerID
	p.UserRole = "Owner"
	p.CreatedBy = ownerID
	p.CreatedDate = time.Now()
	p.UpdatedBy = ownerID
	p.UpdatedDate = time.Now()

	return p
}
