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

//Event represents an event on the database
type Event struct {
	EventID             string      `json:"eventId" validate:"required"`
	Title               Translation `json:"title" validate:"required"`
	Description         Translation `json:"description"`
	Active              bool        `json:"active"`
	MainCategoryID      string      `json:"mainCategoryId"`
	SecondaryCategoryID string      `json:"secondaryCategoryId"`
	CountryID           string      `json:"countryId"`
	RegionID            string      `json:"regionId"`
	CityID              string      `json:"cityId"`
	Address             string      `json:"address"`
	Schedules           []Schedule  `json:"schedules"`
	Audit               *Audit      `json:"audit"`
}

//GetAll returns all events on the system
func (e *Event) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterExpression, filterValues := common.ParseRequestFilters(request)
	result, err := db.Scan(common.EventsTable, filterExpression, filterValues)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []Event{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//Load get Event information from the database
func (e *Event) Load(request events.APIGatewayProxyRequest) error {
	result, err := db.GetItem(common.EventsTable, "eventId", request.PathParameters["id"])
	if err != nil {
		return err
	}
	err = dynamodbattribute.UnmarshalMap(result.Item, e)
	if err != nil {
		return err
	}
	return nil
}

//SaveNew creates a new event on the system
func (e *Event) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), e)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	e.EventID = uuid.New().String()
	e.Active = true
	e.Audit = NewAudit(common.GetTokenUser(request).UserID)
	s := Schedule{}
	e.Schedules = append(e.Schedules, s)

	validate := validator.New()
	err = validate.Struct(e)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	err = db.PutItem(e, common.EventsTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	// Because of a problem with the dynamodb sdk need to create a dummy object and delete to get an empty list
	err = db.DeleteListItem(common.EventsTable, "eventId", e.EventID, "schedules", 0)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	e.Schedules = []Schedule{}
	return common.APIResponse(e, http.StatusCreated)
}

//Update modify an event attributes
func (e *Event) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	//TODO Check if request body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	result, err := db.UpdateItem(common.EventsTable, "eventId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	dynamodbattribute.UnmarshalMap(result.Attributes, e)
	return common.APIResponse(e, http.StatusOK)
}

//Delete remove an event
func (e *Event) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO register in audit table this action
	//TODO implement marked to delete

	err := db.DeleteItem(common.EventsTable, "eventId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

//UpdateEventAudit updates only the audit attributes of a Event using the request information
func UpdateEventAudit(request events.APIGatewayProxyRequest) error {
	jsonMap := make(map[string]interface{})
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	_, err := db.UpdateItem(common.EventsTable, "eventId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return err
	}

	return nil
}
