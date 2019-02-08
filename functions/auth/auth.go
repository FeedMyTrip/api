package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources/auth"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	authentication := auth.Auth{}
	switch req.Resource {
	case "/auth/login":
		switch req.HTTPMethod {
		case "POST":
			return authentication.Login(req)
		}
	case "/auth/refresh":
		switch req.HTTPMethod {
		case "GET":
			return authentication.Refresh(req)
		}
	case "/auth/register":
		switch req.HTTPMethod {
		case "POST":
			return authentication.Register(req)
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
