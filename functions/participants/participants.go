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
	case "POST":
		return postParticipants(req)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
		}, nil
	}
}

func postParticipants(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	participant, err := resources.NewParticipant(req.Body)
	if err != nil {
		return common.APIError(http.StatusUnprocessableEntity, err)
	}

	jsonMap := make(map[string]interface{})
	participants := []resources.Participant{}
	participants = append(participants, *participant)
	jsonMap[":participants"] = participants

	result, err := db.PutListItem("Trips", "tripId", req.PathParameters["id"], "participants", jsonMap)
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

func main() {
	lambda.Start(router)
}
