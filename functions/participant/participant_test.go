package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestPatchParticipantRoute(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PATCH",
		Body: `{
			"userRole": "Editor"
		}`,
		PathParameters: map[string]string{
			"id":            "bf298ece-7b3f-42d4-8970-e5f5649b7ae9",
			"participantId": "b4fba3b2-e21b-4a0e-b45c-6f60fac15e5f",
		},
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestDeleteParticipantRoute(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		PathParameters: map[string]string{
			"id":            "bf298ece-7b3f-42d4-8970-e5f5649b7ae9",
			"participantId": "b4fba3b2-e21b-4a0e-b45c-6f60fac15e5f",
		},
	}

	response, err := router(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
