package common

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

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
