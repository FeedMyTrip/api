package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/feedmytrip/api/resources"

	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.PathParameters["id"] == "" {
		return common.APIError(http.StatusBadRequest, errors.New("invalid path, lack of id parameter"))
	}

	switch req.HTTPMethod {
	case "PATCH":
		return updateTrip(req)
	case "DELETE":
		return deleteTrip(req)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
		}, nil
	}
}

func updateTrip(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(req.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	result, err := db.UpdateItem("Trips", "tripId", req.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	trip := resources.Trip{}
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &trip)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	jsonTrip, err := json.Marshal(trip)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonTrip),
	}, nil
}

func deleteTrip(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	err := db.DeleteItem("Trips", "tripId", req.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(router)
}
