package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	token        string
	loggedUserID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.UsersTable = "UsersTest"
	common.TripsTable = "TripsTest"
	db.CreateTable("UsersTest", "userId", 1, 1)

	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	user, _ := resources.LoginUser(credentials)
	suite.token = *user.Tokens.AccessToken
	suite.loggedUserID = user.UserID
}

func (suite *FeedMyTripAPITestSuite) Test0010GetUserDetails() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": {
				"en": "FMT - Testing Trip #1"
			}
		}`,
	}
	trip := resources.Trip{}
	response, err := trip.SaveNew(req)

	req = events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": {
				"en": "FMT - Testing Trip #2"
			}
		}`,
	}
	trip = resources.Trip{}
	response, err = trip.SaveNew(req)

	req = events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id": suite.loggedUserID,
		},
		QueryStringParameters: map[string]string{
			"include": "trips",
		},
	}

	user := resources.User{}
	response, err = user.GetUserDetails(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
