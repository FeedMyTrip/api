package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	ItineraryID string      `json:"itineraryID" validate:"required"`
	Title       Translation `json:"title" validate:"required"`
	OwnerID     string      `json:"ownerId" validate:"required"`
	StartDate   time.Time   `json:"startDate" validate:"required"`
	EndDate     time.Time   `json:"endDate" validate:"required"`
	Events      []UserEvent `json:"events"`
	Audit       *Audit      `json:"audit"`
}

//ItineraryResponse returns the newly created Itinerary with the tripId
type ItineraryResponse struct {
	TripID    string     `json:"tripId" validate:"required"`
	Itinerary *Itinerary `json:"itinerary" validate:"required"`
}

//SaveNew creates a new Itinerary to a Trip on the database
func (i *Itinerary) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	loggedUser := common.GetTokenUser(request)

	t := Trip{}
	t.Load(request.PathParameters["id"])

	if !t.IsTripAdmin(loggedUser.UserID) && !t.IsTripEditor(loggedUser.UserID) {
		return common.APIError(http.StatusBadRequest, errors.New("itinerary creation not authorized"))
	}

	err := json.Unmarshal([]byte(request.Body), i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	i.Audit = NewAudit(loggedUser.UserID)
	i.OwnerID = loggedUser.UserID
	i.ItineraryID = uuid.New().String()
	event := UserEvent{}
	i.Events = append(i.Events, event)

	validate := validator.New()
	err = validate.Struct(i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap := make(map[string]interface{})
	itineraries := []Itinerary{}
	itineraries = append(itineraries, *i)
	jsonMap[":itineraries"] = itineraries

	result, err := db.PutListItem(common.TripsTable, "tripId", request.PathParameters["id"], "itineraries", jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	t = Trip{}
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &t)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	index := getItineraryIndex(t.Itineraries, i.ItineraryID)
	if index == -1 {
		return common.APIError(http.StatusNotFound, errors.New("itinerary not found"))
	}

	// Because of a problem with the dynamodb sdk need to create a dummy event and delete to get an empty list
	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "itineraries["+strconv.Itoa(index)+"].events", 0)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = AddTripToUser(i.OwnerID, t.TripID, UserTripEditScope)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	ri := ItineraryResponse{}
	ri.TripID = t.TripID
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
	delete(jsonMap, "itineraryID")
	delete(jsonMap, "ownerId")
	delete(jsonMap, "audit.createdBy")
	delete(jsonMap, "audit.createdDate")

	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	t := Trip{}
	t.Load(request.PathParameters["id"])
	index := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if index == -1 {
		return common.APIError(http.StatusNotFound, errors.New("itinerary not found"))
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
	t.Load(request.PathParameters["id"])
	index := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if index == -1 {
		return common.APIError(http.StatusNotFound, errors.New("itinerary not found"))
	}

	if t.ItineraryID == t.Itineraries[index].ItineraryID {
		return common.APIError(http.StatusBadRequest, errors.New("itinerary marked as principal can't be deleted"))
	}

	err := db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "itineraries", index)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	deletedItineraryOwnerID := t.Itineraries[index].OwnerID

	if !t.IsTripAdmin(deletedItineraryOwnerID) {
		total := 0
		for _, i := range t.Itineraries {
			if i.OwnerID == deletedItineraryOwnerID {
				total++
			}
		}
		if total < 2 {
			err = AddTripToUser(deletedItineraryOwnerID, t.TripID, UserTripViewScope)
			if err != nil {
				return common.APIError(http.StatusInternalServerError, err)
			}
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

//NewDefaultItinerary returns a itinerary with the default attributes
func NewDefaultItinerary(userID string) *Itinerary {
	i := &Itinerary{}
	i.ItineraryID = uuid.New().String()
	i.Title.EN = "Default"
	i.Title.PT = "Padrão"
	i.Title.ES = "Estándar"
	i.OwnerID = userID
	i.StartDate = time.Now()
	i.EndDate = i.StartDate.AddDate(0, 0, 15)
	event := UserEvent{}
	i.Events = append(i.Events, event)
	i.Audit = NewAudit(userID)
	return i
}

func getItineraryIndex(itineraries []Itinerary, itineraryID string) int {
	for index, p := range itineraries {
		if p.ItineraryID == itineraryID {
			return index
		}
	}
	return -1
}
