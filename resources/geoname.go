package resources

import (
	"encoding/json"
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
	GeonameID   string      `json:"geonameId" validate:"required"`
	CountryID   string      `json:"countryId"`
	Translation Translation `json:"translation" validate:"required"`
}

//SaveNew creates a new country or city
func (g *Geoname) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), g)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	g.GeonameID = uuid.New().String()
	if g.CountryID == "" {
		g.CountryID = "none"
	}
	validate := validator.New()
	err = validate.Struct(g)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	err = db.PutItem(g, common.GeonamesTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(g, http.StatusCreated)
}

//GetAll returns all countries and cities available in the database
func (g *Geoname) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
