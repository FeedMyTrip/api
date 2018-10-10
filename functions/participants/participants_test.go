package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestPostParticipantsRoute(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Body: `{
			"userId": "000005",
			"userRole": "Viewer"
		}`,
		PathParameters: map[string]string{
			"id": "bf298ece-7b3f-42d4-8970-e5f5649b7ae9",
		},
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
