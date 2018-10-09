package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return getTrips(req)
	case "POST":
		return postTrip(req)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
		}, nil
	}
}

func postTrip(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	trip, err := resources.NewTrip(req.Body)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	err = db.PutItem(trip, "Trips")
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	jsonTrip, err := json.Marshal(trip)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(jsonTrip),
	}, nil
}

func getTrips(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	result, err := db.GetAllItems("Trips")
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	trips := []resources.Trip{}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &trips)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	jsonTrips, err := json.Marshal(trips)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonTrips),
	}, nil
}

func main() {
	lambda.Start(router)
}
