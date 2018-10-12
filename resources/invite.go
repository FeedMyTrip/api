package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

// Invite represents an email to invite a new system user to be a participant of the Trip
type Invite struct {
	InviteID    string    `json:"inviteId" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	CreatedBy   string    `json:"createdBy"`
	CreatedDate time.Time `json:"createdDate"`
}

//InviteResponse returns the Invite with the tripId
type InviteResponse struct {
	TripID string  `json:"tripId" validate:"required"`
	Invite *Invite `json:"invite" validate:"required"`
}

//SaveNew creates a new user invite for the Trip on the database
func (i *Invite) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	//TODO replace 000001 by the userID that execute the action from Cognito
	i.InviteID = uuid.New().String()
	i.CreatedBy = "000001"
	i.CreatedDate = time.Now()

	validate := validator.New()
	err = validate.Struct(i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap := make(map[string]interface{})
	invites := []Invite{}
	invites = append(invites, *i)
	jsonMap[":invites"] = invites

	_, err = db.PutListItem("Trips", "tripId", request.PathParameters["id"], "invites", jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	//TODO Send an email with the invite

	ir := InviteResponse{}
	ir.TripID = request.PathParameters["id"]
	ir.Invite = i

	return common.APIResponse(ir, http.StatusCreated)
}

//Delete remove the Trip invite from the database
func (i *Invite) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	t := Trip{}
	t.LoadTrip(request)
	index, err := getInviteIndex(t.Invites, request.PathParameters["inviteId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	err = db.DeleteListItem("Trips", "tripId", t.TripID, "invites", index)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func getInviteIndex(invites []Invite, inviteID string) (int, error) {
	index := 0
	found := false

	for _, i := range invites {
		if i.InviteID == inviteID {
			found = true
			break
		}
		index++
	}

	if !found {
		return -1, errors.New("invite not found")
	}

	return index, nil
}
