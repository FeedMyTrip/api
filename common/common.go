package common

import "github.com/aws/aws-lambda-go/events"

//APIError generates an api error message response with the defines error and status code
func APIError(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       err.Error(),
	}, nil
}
