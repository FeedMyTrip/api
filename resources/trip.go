package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"

	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"

	validator "gopkg.in/go-playground/validator.v9"
)

// Trip represents a user trip
type Trip struct {
	TripID       string        `json:"tripId" validate:"required"`
	Title        Translation   `json:"title" validate:"required"`
	Description  Translation   `json:"description"`
	Mode         string        `json:"Mode"`
	ItineraryID  string        `json:"itineraryID"`
	Like         int           `json:"like"`
	Dislike      int           `json:"dislike"`
	Itineraries  []Itinerary   `json:"itineraries"`
	Participants []Participant `json:"participants"`
	Invites      []Invite      `json:"invites"`
	Audit        *Audit        `json:"audit"`
}

//IsTripAdmin validates if a user is a Trip admin
func (t *Trip) IsTripAdmin(userID string) bool {
	for _, p := range t.Participants {
		if p.UserID == userID && p.UserRole == "Admin" {
			return true
		}
	}
	return false
}

//GetAll returns all Trips the user can view
func (t *Trip) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterExpression, filterValues := common.ParseRequestFilters(request)
	result, err := db.Scan(common.TripsTable, filterExpression, filterValues)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []Trip{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//Load get trip information from the database
func (t *Trip) Load(id string) error {
	tripResult, err := db.GetItem(common.TripsTable, "tripId", id)
	if err != nil {
		return err
	}
	err = dynamodbattribute.UnmarshalMap(tripResult.Item, t)
	if err != nil {
		return err
	}
	return nil
}

//SaveNew creates a new Trip on the database
func (t *Trip) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), t)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	userID := common.GetTokenUser(request).UserID

	t.TripID = uuid.New().String()
	t.Itineraries = append(t.Itineraries, *NewDefaultItinerary(userID))
	t.ItineraryID = t.Itineraries[0].ItineraryID

	t.Participants = append(t.Participants, *NewOwner(userID))

	i := Invite{}
	t.Invites = append(t.Invites, i)

	t.Audit = NewAudit(userID)

	validate := validator.New()
	err = validate.Struct(t)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if t.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("can't create Trip without Title"))
	}

	err = db.PutItem(t, common.TripsTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	AddTripToUser(userID, t.TripID)

	// Because of a problem with the dynamodb sdk need to create a dummy invite and delete to get an empty list
	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "invites", 0)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	// Because of a problem with the dynamodb sdk need to create a dummy event and delete to get an empty list
	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "itineraries[0].events", 0)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	t.Invites = []Invite{}
	t.Itineraries[0].Events = []UserEvent{}
	return common.APIResponse(t, http.StatusCreated)
}

//Update saves modified attributes to the database using the request information
func (t *Trip) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//TODO Check if request body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	result, err := db.UpdateItem(common.TripsTable, "tripId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	dynamodbattribute.UnmarshalMap(result.Attributes, t)
	return common.APIResponse(t, http.StatusOK)
}

//UpdateTripAudit updates only the audit attributes of a Trip using the request information
func UpdateTripAudit(request events.APIGatewayProxyRequest) error {
	jsonMap := make(map[string]interface{})
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	_, err := db.UpdateItem(common.TripsTable, "tripId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return err
	}
	return nil
}

//Delete remove the Trip from the database
func (t *Trip) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	//TODO validate if user is admin or trip owner
	//TODO register in audit table this action
	//TODO implement marked to delete

	t.Load(request.PathParameters["id"])

	err := db.DeleteItem(common.TripsTable, "tripId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	for _, p := range t.Participants {
		RemoveTripFromUser(p.UserID, t.TripID)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
