package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestPatchTripRoute(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PATCH",
		Body: `{
			"title": "Golang Test"
		}`,
		PathParameters: map[string]string{
			"id": "c638d7ff-a8c8-4866-9a59-77e00c1f05f8",
		},
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
