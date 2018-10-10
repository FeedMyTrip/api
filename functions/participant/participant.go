package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "PATCH":
		return updateParticipant(req)
	case "DELETE":
		return deleteParticipant(req)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
		}, nil
	}
}

func deleteParticipant(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
	}, nil
}

func updateParticipant(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(req.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	//TODO change to user id that executes the action
	jsonMap["updatedBy"] = "000002"
	jsonMap["updatedDate"] = time.Now()

	tripResult, err := db.GetItem("Trips", "tripId", req.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	trip := resources.Trip{}
	err = dynamodbattribute.UnmarshalMap(tripResult.Item, &trip)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	index := 0
	participantFound := false

	for _, p := range trip.Participants {
		if p.ParticipantID == req.PathParameters["participantId"] {
			participantFound = true
			break
		}
		index++
	}

	if !participantFound {
		return common.APIError(http.StatusNotFound, errors.New("Invalid participantId"))
	}

	result, err := db.UpdateListItem("Trips", "tripId", req.PathParameters["id"], "participants", index, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	trip = resources.Trip{}
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &trip)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	jsonTrip, err := json.Marshal(trip)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       string(jsonTrip),
	}, nil
}

func main() {
	lambda.Start(router)
}
