package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return getTrips(req)
	case "POST":
		return postTrip(req)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
		}, nil
	}
}

func postTrip(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := `
	{
		"ID": "trp000001",
		"title": "Post >>> Europe 2019 - Italy and France",
		"description": "Vivamus quis semper metus, non tincidunt dolor",
		"imageBanner": "store/image.jgp",
		"imageCard": "store/image.jgp",
		"createdBy": "USR.181409001",
		"createdDate": "2018-09-19T13:35:59-03:00",
		"modifiedBy": "USR.181409001",
		"modifiedDate": "2018-09-19T13:35:59-03:00"
	}
	`

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       body,
	}, nil
}

func getTrips(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := `
	[
		{
			"ID": "trp000001",
			"title": "Get >>> Europe 2019 - Italy and France",
			"description": "Vivamus quis semper metus, non tincidunt dolor",
			"imageCard": "store/image.jgp",
			"createdBy": "USR.181409001",
			"createdDate": "2018-09-19T13:35:59-03:00",
			"modifiedBy": "USR.181409001",
			"modifiedDate": "2018-09-19T13:35:59-03:00"
		},
		{
			"ID": "trp000002",
			"title": "Get >>> EUA 2020 - New York",
			"description": "Vivamus quis semper metus, non tincidunt dolor",
			"createdBy": "USR.181409001",
			"createdDate": "2018-09-19T13:35:59-03:00",
			"modifiedBy": "USR.181409001",
			"modifiedDate": "2018-09-19T13:35:59-03:00"
		}
  	]
	`
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       body,
	}, nil
}

func main() {
	lambda.Start(router)
}
