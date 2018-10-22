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

//UserEvent represents an user event linked with a trip itinerary
type UserEvent struct {
	UserEventID            string    `json:"userEventId" validate:"required"`
	ItineraryID            string    `json:"itineraryID" validate:"required"`
	LanguageCode           string    `json:"languageCode" validate:"required"`
	Title                  string    `json:"title" validate:"required"`
	TripID                 string    `json:"tripId"`
	EventID                string    `json:"eventId"`
	Description            string    `json:"description"`
	StartDate              time.Time `json:"startDate"`
	EndDate                time.Time `json:"endDate"`
	ItinerarySecondsOffset int       `json:"itinerarySecondsOffset"`
	MainCategoryID         string    `json:"mainCategoryId"`
	SecondaryCategoryID    string    `json:"secondaryCategoryId"`
	CountryID              string    `json:"countryId"`
	CityID                 string    `json:"cityId"`
	Address                string    `json:"address"`
	Evaluated              bool      `json:"evaluated"`
	Audit                  *Audit    `json:"audit"`
}

//GetAll return all user events filtered
func (u *UserEvent) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterExpression, filterValues := common.ParseRequestFilters(request)
	result, err := db.GetAllItems(common.UserEventsTable, filterExpression, filterValues)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []UserEvent{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//SaveNew creates a new user event
func (u *UserEvent) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), u)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	u.UserEventID = uuid.New().String()
	u.Audit = NewAudit(common.GetTokenUser(request))
	u.ItinerarySecondsOffset = -1

	validate := validator.New()
	err = validate.Struct(u)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	err = db.PutItem(u, common.UserEventsTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(u, http.StatusCreated)
}

//Update modify an existent user event
func (u *UserEvent) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//TODO Check if request body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}
	jsonMap["audit.updatedBy"] = "000002"
	jsonMap["audit.updatedDate"] = time.Now()

	result, err := db.UpdateItem(common.UserEventsTable, "userEventId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	dynamodbattribute.UnmarshalMap(result.Attributes, u)
	return common.APIResponse(u, http.StatusOK)
}

//Delete removes an user event
func (u *UserEvent) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	credentialConfirmed := false
	//TODO check if user is a system admin and delete
	//TODO check if user is trip admin
	if credentialConfirmed == false {
		u.LoadUserEvent(request.PathParameters["id"])
		if u.Audit.CreatedBy == common.GetTokenUser(request) {
			credentialConfirmed = true
		}
	}

	if credentialConfirmed == false {
		return common.APIError(http.StatusForbidden, errors.New("permission denied to delete this user event"))
	}

	err := db.DeleteItem(common.UserEventsTable, "userEventId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

//LoadUserEvent get user event attributes from databse
func (u *UserEvent) LoadUserEvent(id string) error {
	response, err := db.GetItem(common.UserEventsTable, "userEventId", id)
	if err != nil {
		return err
	}
	err = dynamodbattribute.UnmarshalMap(response.Item, u)
	if err != nil {
		return err
	}
	return nil
}
