package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources/categories"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	category := categories.Category{}
	switch req.Resource {
	case "/categories":
		switch req.HTTPMethod {
		case "GET":
			return category.GetAll(req)
		case "POST":
			return category.SaveNew(req)
		}
	case "/categories/{id}":
		switch req.HTTPMethod {
		case "DELETE":
			return category.Delete(req)
		case "PATCH":
			return category.Update(req)
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
