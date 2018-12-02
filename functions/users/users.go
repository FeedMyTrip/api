package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources/users"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user := users.User{}
	switch req.Resource {
	case "/users":
		switch req.HTTPMethod {
		case "GET":
			return user.GetAll(req)
		}
	case "/users/{id}":
		switch req.HTTPMethod {
		case "PATCH":
			return user.Update(req)
		case "DELETE":
			return user.Delete(req)
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
