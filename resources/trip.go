package resources

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"

	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"

	validator "gopkg.in/go-playground/validator.v9"
)

// Trip represents a user trip
type Trip struct {
	TripID       string        `json:"tripId" validate:"required"`
	Title        string        `json:"title" validate:"required"`
	Description  string        `json:"description"`
	ItineraryID  string        `json:"itineraryID"`
	Itineraries  []Itinerary   `json:"itineraries"`
	Participants []Participant `json:"participants"`
	Invites      []Invite      `json:"invites"`
	Audit        *Audit        `json:"audit"`
}

//GetAll returns all Trips the user can view
func (t *Trip) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//TODO filter to return only user Trips
	result, err := db.GetAllItems(common.TripsTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	return tripAPIListResponse(result.Items, http.StatusOK)
}

//LoadTrip get trip information from the database
func (t *Trip) LoadTrip(request events.APIGatewayProxyRequest) error {
	tripResult, err := db.GetItem(common.TripsTable, "tripId", request.PathParameters["id"])
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

	//TODO replace 000001 by the userID from Cognito
	tokenUser := "000001"

	t.TripID = uuid.New().String()
	t.Itineraries = append(t.Itineraries, *NewDefaultItinerary(tokenUser))
	t.ItineraryID = t.Itineraries[0].ItineraryID

	t.Participants = append(t.Participants, *NewOwner(tokenUser))

	i := Invite{}
	t.Invites = append(t.Invites, i)

	t.Audit = NewAudit(tokenUser)

	validate := validator.New()
	err = validate.Struct(t)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	err = db.PutItem(t, common.TripsTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	// Because of a problem with the dynamodb sdk need to create a dummy invite and delete to get an empty list
	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "invites", 0)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	jsonTrip, err := json.Marshal(t)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(jsonTrip),
	}, nil
}

//Update saves modified attributes to the database using the request information
func (t *Trip) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//TODO Check if request body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}
	jsonMap["audit.updatedBy"] = "000002"
	jsonMap["audit.updatedDate"] = time.Now()

	result, err := db.UpdateItem(common.TripsTable, "tripId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return tripAPIResponse(result.Attributes, http.StatusOK)
}

//UpdateTripAudit updates only the audit attributes of a Trip using the request information
func UpdateTripAudit(request events.APIGatewayProxyRequest) error {
	jsonMap := make(map[string]interface{})
	jsonMap["audit.updatedBy"] = "000002"
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

	err := db.DeleteItem(common.TripsTable, "tripId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func tripAPIResponse(atributtes map[string]*dynamodb.AttributeValue, statuscode int) (events.APIGatewayProxyResponse, error) {
	t := Trip{}
	err := dynamodbattribute.UnmarshalMap(atributtes, &t)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	return common.APIResponse(t, statuscode)
}

func tripAPIListResponse(items []map[string]*dynamodb.AttributeValue, statuscode int) (events.APIGatewayProxyResponse, error) {
	tripsList := []Trip{}

	err := dynamodbattribute.UnmarshalListOfMaps(items, &tripsList)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(tripsList, statuscode)
}
