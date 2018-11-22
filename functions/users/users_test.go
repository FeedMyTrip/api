package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources/auth"
	"github.com/feedmytrip/api/resources/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	token  string
	userID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	user, _ := auth.LoginUser(credentials)
	suite.token = *user.Tokens.AccessToken
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewUser() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"id": "0001",
			"first_name": "Laura",
			"last_name": "Morgan"
		}`,
	}

	user := users.User{}
	response, err := user.SaveNew(req)
	suite.userID = "0001"

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0020UpdateUser() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"about_me": "texto sobre mim"
		}`,
		PathParameters: map[string]string{
			"id": suite.userID,
		},
	}

	user := users.User{}
	response, err := user.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0030GetAllUsers() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
	}

	user := users.User{}
	response, err := user.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0040DeleteUser() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id": suite.userID,
		},
	}

	user := users.User{}
	response, err := user.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
