package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources/trips"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	switch req.Resource {
	case "/trips":
		trip := trips.Trip{}
		switch req.HTTPMethod {
		case "GET":
			return trip.GetAll(req)
		case "POST":
			return trip.SaveNew(req)
		}
	case "/trips/{id}":
		trip := trips.Trip{}
		switch req.HTTPMethod {
		case "GET":
			return trip.Get(req)
		case "PATCH":
			return trip.Update(req)
		case "DELETE":
			return trip.Delete(req)
		}
	case "/trips/{id}/participants":
		participant := trips.Participant{}
		switch req.HTTPMethod {
		case "GET":
			return participant.GetAll(req)
		case "POST":
			return participant.SaveNew(req)
		}
	case "/trips/{id}/participants/{participant_id}":
		participant := trips.Participant{}
		switch req.HTTPMethod {
		case "PATCH":
			return participant.Update(req)
		case "DELETE":
			return participant.Delete(req)
		}
	case "/trips/{id}/invites":
		invite := trips.Invite{}
		switch req.HTTPMethod {
		case "GET":
			return invite.GetAll(req)
		case "POST":
			return invite.SaveNew(req)
		}
	case "/trips/{id}/invites/{invite_id}":
		invite := trips.Invite{}
		switch req.HTTPMethod {
		case "DELETE":
			return invite.Delete(req)
		}
	case "/trips/{id}/itineraries":
		itinerary := trips.Itinerary{}
		switch req.HTTPMethod {
		case "GET":
			return itinerary.GetAll(req)
		case "POST":
			return itinerary.SaveNew(req)
		}
	case "/trips/{id}/itineraries/{itinerary_id}":
		itinerary := trips.Itinerary{}
		switch req.HTTPMethod {
		case "PATCH":
			return itinerary.Update(req)
		case "DELETE":
			return itinerary.Delete(req)
		}
	case "/trips/{id}/itineraries/{itinerary_id}/events":
		event := trips.ItineraryEvent{}
		switch req.HTTPMethod {
		case "GET":
			return event.GetAll(req)
		case "POST":
			return event.SaveNew(req)
		}
	case "/trips/{id}/itineraries/{itinerary_id}/events/{event_id}":
		event := trips.ItineraryEvent{}
		switch req.HTTPMethod {
		case "PATCH":
			return event.Update(req)
		case "DELETE":
			return event.Delete(req)
		}
	case "/trips/{id}/itineraries/{itinerary_id}/add/{global_event_id}":
		event := trips.ItineraryEvent{}
		switch req.HTTPMethod {
		case "POST":
			return event.Add(req)
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
