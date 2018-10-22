package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	userEvent := resources.UserEvent{}
	switch req.Resource {
	case "/userevents":
		switch req.HTTPMethod {
		case "GET":
			return userEvent.GetAll(req)
		case "POST":
			return userEvent.SaveNew(req)
		}
	case "/userevents/{id}":
		switch req.HTTPMethod {
		case "PATCH":
			return userEvent.Update(req)
		case "DELETE":
			return userEvent.Delete(req)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
	}, nil
}

func main() {
	lambda.Start(router)
}
