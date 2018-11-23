package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	fmt "github.com/feedmytrip/api/resources/events"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.Resource {
	case "/events":
		event := fmt.Event{}
		switch req.HTTPMethod {
		case "GET":
			return event.GetAll(req)
		case "POST":
			return event.SaveNew(req)
		}
	case "/events/{id}":
		event := fmt.Event{}
		switch req.HTTPMethod {
		case "GET":
			return event.Get(req)
		case "DELETE":
			return event.Delete(req)
		case "PATCH":
			return event.Update(req)
		}
	case "/events/{id}/schedules":
		schedule := fmt.Schedule{}
		switch req.HTTPMethod {
		case "POST":
			return schedule.SaveNew(req)
		case "GET":
			return schedule.GetAll(req)
		}
	case "/events/{id}/schedules/{schedule_id}":
		schedule := fmt.Schedule{}
		switch req.HTTPMethod {
		case "PATCH":
			return schedule.Update(req)
		case "DELETE":
			return schedule.Delete(req)
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
