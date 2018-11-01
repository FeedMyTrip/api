package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	refreshToken string
}

func (suite *FeedMyTripAPITestSuite) Test0010Login() {
	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	req := events.APIGatewayProxyRequest{
		Body: credentials,
	}

	auth := resources.Auth{}
	response, err := auth.Login(req)
	user := resources.AuthUserResponse{}
	json.Unmarshal([]byte(response.Body), &user)
	suite.refreshToken = *user.Tokens.RefreshToken

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0020RefreshToken() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.refreshToken,
		},
	}

	auth := resources.Auth{}
	response, err := auth.Refresh(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
