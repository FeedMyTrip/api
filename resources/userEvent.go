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

//UserEvent represents an user event linked with a trip itinerary
type UserEvent struct {
	UserEventID         string      `json:"userEventId" validate:"required"`
	Title               Translation `json:"title" validate:"required"`
	Description         Translation `json:"description"`
	EventID             string      `json:"eventId"`
	BeginOffset         int         `json:"beginOffset"`
	Duration            int         `json:"duration"`
	MainCategoryID      string      `json:"mainCategoryId"`
	SecondaryCategoryID string      `json:"secondaryCategoryId"`
	CountryID           string      `json:"countryId"`
	RegionID            string      `json:"regionId"`
	CityID              string      `json:"cityId"`
	Address             string      `json:"address"`
	Evaluated           bool        `json:"evaluated"`
	Audit               *Audit      `json:"audit"`
}

//UserEventResponse represents a response with the tripId, itineraryId and the created event
type UserEventResponse struct {
	TripID      string     `json:"tripId" validate:"required"`
	ItineraryID string     `json:"itineraryID" validate:"required"`
	Event       *UserEvent `json:"event" validate:"required"`
}

//AddGlobal creates a new user event based on a global event
func (u *UserEvent) AddGlobal(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	global := Event{}
	err := global.Load(request.PathParameters["globalEventId"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	u.Title = global.Title
	u.Description = global.Description
	u.EventID = global.EventID
	u.MainCategoryID = global.MainCategoryID
	u.SecondaryCategoryID = global.SecondaryCategoryID
	u.CountryID = global.CountryID
	u.CityID = global.CityID
	u.RegionID = global.RegionID
	u.Address = global.Address
	u.BeginOffset = -1
	u.Duration = 28800

	return save(u, request)
}

//SaveNew creates a new user event
func (u *UserEvent) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), u)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	u.BeginOffset = -1
	u.Duration = 28800
	u.EventID = ""

	return save(u, request)
}

//Update modify an existent user event
func (u *UserEvent) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	// TODO validate fields before update (userId, ItineraryId and auditing fileds should not be updated)

	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	t := Trip{}
	t.LoadTrip(request.PathParameters["id"])
	itineraryIndex, err := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	eventIndex, err := getItineraryEventIndex(t.Itineraries[itineraryIndex].Events, request.PathParameters["eventId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	listName := "itineraries[" + strconv.Itoa(itineraryIndex) + "].events"
	result, err := db.UpdateListItem(common.TripsTable, "tripId", t.TripID, listName, eventIndex, jsonMap)
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

	uer := UserEventResponse{}
	uer.TripID = request.PathParameters["id"]
	uer.ItineraryID = request.PathParameters["itineraryId"]
	uer.Event = &t.Itineraries[itineraryIndex].Events[eventIndex]
	return common.APIResponse(uer, http.StatusOK)
}

//Delete removes an user event
func (u *UserEvent) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	t := Trip{}
	t.LoadTrip(request.PathParameters["id"])

	itineraryIndex, err := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	eventIndex, err := getItineraryEventIndex(t.Itineraries[itineraryIndex].Events, request.PathParameters["eventId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	user := common.GetTokenUser(request)
	if !user.IsAdmin() {
		if !t.IsTripAdmin(user.UserID) {
			if t.Itineraries[itineraryIndex].UserID != user.UserID {
				return common.APIError(http.StatusForbidden, errors.New("no permission to delete event"))
			}
		}
	}

	err = db.DeleteListItem(common.TripsTable, "tripId", t.TripID, "itineraries["+strconv.Itoa(itineraryIndex)+"].events", eventIndex)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}

func save(u *UserEvent, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	u.UserEventID = uuid.New().String()
	u.Audit = NewAudit(common.GetTokenUser(request).UserID)

	validate := validator.New()
	err := validate.Struct(u)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if u.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("can't create a new event without a title"))
	}

	t := Trip{}
	t.LoadTrip(request.PathParameters["id"])
	index, err := getItineraryIndex(t.Itineraries, request.PathParameters["itineraryId"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	jsonMap := make(map[string]interface{})
	events := []UserEvent{}
	events = append(events, *u)
	jsonMap[":itineraries_events"] = events

	_, err = db.PutListItem(common.TripsTable, "tripId", request.PathParameters["id"], "itineraries["+strconv.Itoa(index)+"].events", jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateTripAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	uer := UserEventResponse{}
	uer.TripID = request.PathParameters["id"]
	uer.ItineraryID = request.PathParameters["itineraryId"]
	uer.Event = u
	return common.APIResponse(uer, http.StatusCreated)
}

func getItineraryEventIndex(events []UserEvent, eventID string) (int, error) {
	index := 0
	found := false

	for _, e := range events {
		if e.UserEventID == eventID {
			found = true
			break
		}
		index++
	}

	if !found {
		return -1, errors.New("event not found")
	}

	return index, nil
}
