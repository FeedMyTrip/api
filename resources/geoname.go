package resources

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

//Geoname represents a country or a city in the system
type Geoname struct {
	GeonameID string      `json:"geonameId" validate:"required"`
	CountryID string      `json:"countryId"`
	RegionID  string      `json:"regionId"`
	Title     Translation `json:"title" validate:"required"`
}

//SaveNew creates a new country or city
func (g *Geoname) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), g)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	g.GeonameID = uuid.New().String()
	if g.CountryID == "" {
		g.CountryID = "none"
	}
	if g.RegionID == "" {
		g.RegionID = "none"
	}
	validate := validator.New()
	err = validate.Struct(g)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	g.Title.Translate()

	err = db.PutItem(g, common.GeonamesTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(g, http.StatusCreated)
}

//GetAll returns all countries, regions and cities available in the database
func (g *Geoname) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterExpression, filterValues := common.ParseRequestFilters(request)
	result, err := db.Scan(common.GeonamesTable, filterExpression, filterValues)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []Geoname{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//GetAllByCountry returns all locations from a defined coutry
func (g *Geoname) GetAllByCountry(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := "none"
	if val, ok := request.PathParameters["id"]; ok {
		id = val
	}
	result, err := db.Query(common.GeonamesTable, "countryId", id)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []Geoname{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//Update modify geoname attributes
func (g *Geoname) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO check if body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	result, err := db.UpdateItem(common.GeonamesTable, "geonameId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	dynamodbattribute.UnmarshalMap(result.Attributes, g)
	return common.APIResponse(g, http.StatusOK)
}

//Delete removes a country or city from the database
func (g *Geoname) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO implement mark to delete
	//TODO verify if there is any event with this geoname
	err := db.DeleteItem(common.GeonamesTable, "geonameId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
