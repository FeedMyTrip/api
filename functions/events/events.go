package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.Resource {
	case "/events":
		event := resources.Event{}
		switch req.HTTPMethod {
		case "GET":
			return event.GetAll(req)
		case "POST":
			return event.SaveNew(req)
		}
	case "/events/{id}":
		event := resources.Event{}
		switch req.HTTPMethod {
		case "DELETE":
			return event.Delete(req)
		case "PATCH":
			return event.Update(req)
		}
	case "/events/{id}/translations":
		eventTranslation := resources.EventTranslation{}
		switch req.HTTPMethod {
		case "PUT":
			return eventTranslation.Save(req)
		}
	case "/events/{id}/schedules":
		schedule := resources.Schedule{}
		switch req.HTTPMethod {
		case "POST":
			return schedule.SaveNew(req)
		}
	case "/events/{id}/schedules/{scheduleId}":
		schedule := resources.Schedule{}
		switch req.HTTPMethod {
		case "PATCH":
			return schedule.Update(req)
		case "DELETE":
			return schedule.Delete(req)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
	}, nil
}

func main() {
	lambda.Start(router)
}
