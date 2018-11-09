package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	switch req.Resource {
	case "/trips":
		trip := resources.Trip{}
		switch req.HTTPMethod {
		case "GET":
			return trip.GetAll(req)
		case "POST":
			return trip.SaveNew(req)
		}
	case "/trips/{id}":
		trip := resources.Trip{}
		switch req.HTTPMethod {
		case "PATCH":
			return trip.Update(req)
		case "DELETE":
			return trip.Delete(req)
		}
	case "/trips/{id}/participants":
		participant := resources.Participant{}
		switch req.HTTPMethod {
		case "POST":
			return participant.SaveNew(req)
		}
	case "/trips/{id}/participants/{participantId}":
		participant := resources.Participant{}
		switch req.HTTPMethod {
		case "PATCH":
			return participant.Update(req)
		case "DELETE":
			return participant.Delete(req)
		}
	case "/trips/{id}/invites":
		invite := resources.Invite{}
		switch req.HTTPMethod {
		case "POST":
			return invite.SaveNew(req)
		}
	case "/trips/{id}/invites/{inviteId}":
		invite := resources.Invite{}
		switch req.HTTPMethod {
		case "DELETE":
			return invite.Delete(req)
		}
	case "/trips/{id}/itineraries":
		itinerary := resources.Itinerary{}
		switch req.HTTPMethod {
		case "POST":
			return itinerary.SaveNew(req)
		}
	case "/trips/{id}/itineraries/{itineraryId}":
		itinerary := resources.Itinerary{}
		switch req.HTTPMethod {
		case "PATCH":
			return itinerary.Update(req)
		case "DELETE":
			return itinerary.Delete(req)
		}
	case "/trips/{id}/itineraries/{itineraryId}/events":
		event := resources.UserEvent{}
		switch req.HTTPMethod {
		case "POST":
			return event.SaveNew(req)
		}
	case "/trips/{id}/itineraries/{itineraryId}/events/{eventId}":
		event := resources.UserEvent{}
		switch req.HTTPMethod {
		case "PATCH":
			return event.Update(req)
		case "DELETE":
			return event.Delete(req)
		}
	case "/trips/{id}/itineraries/{itineraryId}/add/{globalEventId}":
		event := resources.UserEvent{}
		switch req.HTTPMethod {
		case "POST":
			return event.AddGlobal(req)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}, nil
}

func main() {
	lambda.Start(router)
}
