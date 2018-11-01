package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"

	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

const (
	//ParticipantOwnerRole defines a participant as a trip owner
	ParticipantOwnerRole = "Owner"
	//ParticipantAdminRole defines a participant as a trip admin
	ParticipantAdminRole = "Admin"
	//ParticipantEditorRole defines a participant as a trip editor
	ParticipantEditorRole = "Editor"
	//ParticipantViewerRole defines a participant as a trip viewer
	ParticipantViewerRole = "Viewer"
)

// Participant represents a user that is participating in the trip
type Participant struct {
	ParticipantID string `json:"participantId" validate:"required"`
	UserID        string `json:"userId" validate:"required"`
	UserRole      string `json:"userRole" validate:"required"`
	Audit         *Audit `json:"audit"`
}

//ParticipantResponse returns the newly created participant with the tripId
type ParticipantResponse struct {
	TripID      string       `json:"tripId" validate:"required"`
	Participant *Participant `json:"participant" validate:"required"`
}

//SaveNew creates a new Participant to a Trip on the database
func (p *Participant) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Check if user Trip role is Admin or Owner to be able to include new participants

	err := json.Unmarshal([]byte(request.Body), p)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	p.ParticipantID = uuid.New().String()
	p.Audit = NewAudit(common.GetTokenUser(request).UserID)
	validate := validator.New()
	err = validate.Struct(p)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap := make(map[string]interface{})
	participants := []Participant{}
	participants = append(participants, *p)
	jsonMap[":participants"] = participants

	_, err = db.PutListItem(common.TripsTable, "tripId", request.PathParameters["id"], "participants", jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	scope := UserTripViewScope
	if p.UserRole == ParticipantAdminRole {
		scope = UserTripEditScope
	}
	err = AddTripToUser(p.UserID, request.PathParameters["id"], scope)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	rp := ParticipantResponse{}
	rp.TripID = request.PathParameters["id"]
	rp.Participant = p
	return common.APIResponse(rp, http.StatusCreated)
}

//Update saves participant modifications to the database
func (p *Participant) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	t := Trip{}
	t.Load(request.PathParameters["id"])
	index, err := getParticipantIndex(t.Participants, request.PathParameters["participantId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	if t.Participants[index].UserRole == ParticipantOwnerRole {
		return common.APIError(http.StatusBadRequest, errors.New("trip owner`s role can not be changed"))
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(request.Body), &p)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap["userRole"] = p.UserRole
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	oldUserRole := t.Participants[index].UserRole
	if oldUserRole == p.UserRole {

	}

	result, err := db.UpdateListItem(common.TripsTable, "tripId", t.TripID, "participants", index, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = dynamodbattribute.UnmarshalMap(result.Attributes, &t)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	scope := ""
	if t.Participants[index].UserRole == ParticipantAdminRole {
		scope = UserTripEditScope
	} else if t.Participants[index].UserRole == ParticipantViewerRole {
		scope = UserTripViewScope
	}

	if scope != "" {
		err = AddTripToUser(p.UserID, request.PathParameters["id"], scope)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
	}

	rp := ParticipantResponse{}
	rp.TripID = request.PathParameters["id"]
	rp.Participant = &t.Participants[index]
	return common.APIResponse(rp, http.StatusOK)
}

//Delete remove a participant from the database
func (p *Participant) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	t := Trip{}
	t.Load(request.PathParameters["id"])
	index, err := getParticipantIndex(t.Participants, request.PathParameters["participantId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	if t.Participants[index].UserRole == ParticipantOwnerRole {
		return common.APIError(http.StatusBadRequest, errors.New("trip owner can not be deleted"))
	}

	deletedUserID := t.Participants[index].UserID

	loggedUser := common.GetTokenUser(request)
	if !loggedUser.IsAdmin() && loggedUser.UserID != deletedUserID && !t.IsTripAdmin(loggedUser.UserID) {
		return common.APIError(http.StatusBadRequest, errors.New("not authorized to delete this participant"))
	}

	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "participants", index)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = RemoveTripFromUser(deletedUserID, t.TripID)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

//NewOwner returns the owner as a participant in the Trip
func NewOwner(userID string) *Participant {
	p := &Participant{}
	p.UserID = userID
	p.UserRole = ParticipantOwnerRole
	p.ParticipantID = uuid.New().String()
	p.Audit = NewAudit(userID)
	return p
}

func getParticipantIndex(participants []Participant, participantID string) (int, error) {
	index := 0
	found := false

	for _, p := range participants {
		if p.ParticipantID == participantID {
			found = true
			break
		}
		index++
	}

	if !found {
		return -1, errors.New("participant not found")
	}

	return index, nil
}
