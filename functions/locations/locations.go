package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources/locations"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	location := locations.Location{}
	switch req.Resource {
	case "/locations":
		switch req.HTTPMethod {
		case "POST":
			return location.SaveNew(req)
		case "GET":
			return location.GetAll(req)
		}
	case "/locations/{id}":
		switch req.HTTPMethod {
		case "DELETE":
			return location.Delete(req)
		case "PATCH":
			return location.Update(req)
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
