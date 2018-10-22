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

// Itinerary represents a way to group trip events
type Itinerary struct {
	ItineraryID string    `json:"itineraryID" validate:"required"`
	Title       string    `json:"title" validate:"required"`
	UserID      string    `json:"userId" validate:"required"`
	StartDate   time.Time `json:"startDate" validate:"required"`
	EndDate     time.Time `json:"endDate" validate:"required"`
	Audit       *Audit    `json:"audit"`
}

//ItineraryResponse returns the newly created Itinerary with the tripId
type ItineraryResponse struct {
	TripID    string     `json:"tripId" validate:"required"`
	Itinerary *Itinerary `json:"itinerary" validate:"required"`
}

//SaveNew creates a new Itinerary to a Trip on the database
func (i *Itinerary) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Check if user Trip role is Admin or Owner to be able to include new itineraries

	err := json.Unmarshal([]byte(request.Body), i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	//TODO replace 000001 by the userID that execute the action from Cognito
	i.Audit = NewAudit("000001")
	i.UserID = "000001"
	i.ItineraryID = uuid.New().String()
	validate := validator.New()
	err = validate.Struct(i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap := make(map[string]interface{})
	itineraries := []Itinerary{}
	itineraries = append(itineraries, *i)
	jsonMap[":itineraries"] = itineraries

	_, err = db.PutListItem(common.TripsTable, "tripId", request.PathParameters["id"], "itineraries", jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	ri := ItineraryResponse{}
	ri.TripID = request.PathParameters["id"]
	ri.Itinerary = i
	return common.APIResponse(ri, http.StatusCreated)
}

//Update saves itinerary modifications to the database
func (i *Itinerary) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	// TODO validate fields before update (userId, ItineraryId and auditing fileds should not be updated)

	//TODO change to user id that executes the action
	jsonMap["audit.updatedBy"] = "000002"
	jsonMap["audit.updatedDate"] = time.Now()

	t := Trip{}
	t.LoadTrip(request.PathParameters["id"])
	index, err := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	result, err := db.UpdateListItem(common.TripsTable, "tripId", t.TripID, "itineraries", index, jsonMap)
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

	ir := ItineraryResponse{}
	ir.TripID = request.PathParameters["id"]
	ir.Itinerary = &t.Itineraries[index]
	return common.APIResponse(ir, http.StatusOK)
}

//Delete remove an itinerary from the database
func (i *Itinerary) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//TODO implement mark to delete
	t := Trip{}
	t.LoadTrip(request.PathParameters["id"])
	index, err := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	if t.ItineraryID == t.Itineraries[index].ItineraryID {
		return common.APIError(http.StatusBadRequest, errors.New("itinerary marked as principal can't be deleted"))
	}

	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "itineraries", index)
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

//NewDefaultItinerary returns a itinerary with the default attributes
func NewDefaultItinerary(userID string) *Itinerary {
	i := &Itinerary{}
	i.ItineraryID = uuid.New().String()
	i.Title = "Default Itinerary"
	i.UserID = userID
	i.StartDate = time.Now()
	i.EndDate = i.StartDate.AddDate(0, 0, 15)
	i.Audit = NewAudit(userID)
	return i
}

func getItineraryIndex(itineraries []Itinerary, itineraryID string) (int, error) {
	index := 0
	found := false

	for _, p := range itineraries {
		if p.ItineraryID == itineraryID {
			found = true
			break
		}
		index++
	}

	if !found {
		return -1, errors.New("Itinerary not found")
	}

	return index, nil
}
