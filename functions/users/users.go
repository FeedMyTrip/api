package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user := resources.User{}
	switch req.Resource {
	case "/users/{id}":
		switch req.HTTPMethod {
		case "GET":
			return user.GetUserDetails(req)
		}
	case "/users/favorites/{contentType}/{contentId}":
		switch req.HTTPMethod {
		case "POST":
			return user.ToggleFavoriteContent(req)
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
