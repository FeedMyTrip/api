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

//Highlight represents a group of trips and events to emphasize
type Highlight struct {
	HighlightID  string      `json:"highlightId"`
	Title        Translation `json:"title"`
	Description  Translation `json:"description"`
	ImagePath    string      `json:"imagePath"`
	Trips        []string    `json:"trips"`
	Events       []string    `json:"events"`
	Active       bool        `json:"active"`
	ScheduleDate time.Time   `json:"scheduleDate"`
	CountryID    string      `json:"countryId"`
	RegionID     string      `json:"regionId"`
	CityID       string      `json:"cityId"`
	Audit        *Audit      `json:"audit"`
}

//SaveNew creates a new highlight
func (h *Highlight) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), h)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	h.HighlightID = uuid.New().String()
	h.Audit = NewAudit(common.GetTokenUser(request).UserID)

	validate := validator.New()
	err = validate.Struct(h)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	err = db.PutItem(h, common.HighlightsTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(h, http.StatusCreated)
}

//GetItem returns a highlight events and trips
func (h *Highlight) GetItem(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//TODO: return highlight events and trips
	return common.APIResponse(h, http.StatusNotImplemented)
}

//GetAll returns all highlights available in the database base on the defined filter
func (h *Highlight) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterExpression, filterValues := common.ParseRequestFilters(request)
	result, err := db.Scan(common.HighlightsTable, filterExpression, filterValues)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []Trip{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//Update modify highlight attributes
func (h *Highlight) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO check if body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	result, err := db.UpdateItem(common.HighlightsTable, "HighlightId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	dynamodbattribute.UnmarshalMap(result.Attributes, h)
	return common.APIResponse(h, http.StatusOK)
}

//Delete removes a highlight from the database
func (h *Highlight) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO implement mark to delete
	err := db.DeleteItem(common.HighlightsTable, "HighlightId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
