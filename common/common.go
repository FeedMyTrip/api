package common

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	//TripsTable defines the databse table to store the Trips and nested objects
	TripsTable = "Trips"
	//EventsTable defines the databse table to store the Events and nested objects
	EventsTable = "Events"
	//UserEventsTable defines the databse table to store the users events and nested objects
	UserEventsTable = "UserEvents"
	//CategoriesTable defines the databse table to store the Categories and nested objects
	CategoriesTable = "Categories"
)

//ParseRequestFilters process the request to parse the querystrings to dynamodb filters
func ParseRequestFilters(request events.APIGatewayProxyRequest) (string, map[string]*dynamodb.AttributeValue) {
	if len(request.QueryStringParameters) == 0 {
		return "", nil
	}
	filterExpression := ""
	data := make(map[string]interface{})
	for k, v := range request.QueryStringParameters {
		if k == "state" && v == "active" {
			filterExpression += "active = :active"
			data[":active"] = true
		} else if k == "limit" {
			data["limit"] = v
		} else {
			filterExpression += k + " = :" + k
			i, err := strconv.Atoi(v)
			if err == nil {
				data[":"+k] = i
			} else {
				data[":"+k] = v
			}
		}
		filterExpression += ", "
	}

	filterExpression = filterExpression[:len(filterExpression)-2]
	filterValues, _ := dynamodbattribute.MarshalMap(data)
	return filterExpression, filterValues
}

//GetTokenUser return the userId from the request token
func GetTokenUser(request events.APIGatewayProxyRequest) string {
	//TODO get user id from AWS Cognito tokengo
	return "0000001"
}

//APIError generates an api error message response with the defines error and status code
func APIError(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	jsonBody := `{"error":"` + err.Error() + `"}`
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       jsonBody,
	}, nil
}

//APIResponse gernerates an APIGatewayProxyResponse based on the interface object
func APIResponse(object interface{}, statuscode int) (events.APIGatewayProxyResponse, error) {
	jsonObject, err := json.Marshal(object)
	if err != nil {
		return APIError(http.StatusUnprocessableEntity, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statuscode,
		Body:       string(jsonObject),
	}, nil
}
