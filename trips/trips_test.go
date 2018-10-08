package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestGetTripRouter(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestPostTripRouter(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestMethodNotAllowedRouter(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "Invalid_Method",
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
}
