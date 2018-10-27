package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/feedmytrip/api/resources"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	highlight := resources.Highlight{}
	switch req.Resource {
	case "/highlights":
		switch req.HTTPMethod {
		case "POST":
			return highlight.SaveNew(req)
		case "GET":
			return highlight.GetAll(req)
		}
	case "/highlights/{id}":
		switch req.HTTPMethod {
		case "GET":
			return highlight.GetItem(req)
		case "DELETE":
			return highlight.Delete(req)
		case "PATCH":
			return highlight.Update(req)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
	}, nil
}

func main() {
	lambda.Start(router)
}
