package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
}

func (suite *FeedMyTripAPITestSuite) Test0040DeleteCategory() {
	req := events.APIGatewayProxyRequest{
		Body: `{
			"username": "test",
			"password": "test12345"
		}`,
	}

	auth := resources.Auth{}
	response, err := auth.Login(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
