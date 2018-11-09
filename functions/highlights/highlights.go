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
	case "/highlights/{id}/add/{contentType}/{contentId}":
		switch req.HTTPMethod {
		case "POST":
			return highlight.AddContent(req)
		}
	case "/highlights/{id}/remove/{contentType}/{contentId}":
		switch req.HTTPMethod {
		case "DELETE":
			return highlight.RemoveContent(req)
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
