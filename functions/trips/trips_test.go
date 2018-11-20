package main

// Basic imports
import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources/auth"
	"github.com/feedmytrip/api/resources/trips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	adminToken       string
	participantToken string
	participantID    string
	tripID           string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	user, _ := auth.LoginUser(credentials)
	suite.adminToken = *user.Tokens.AccessToken

	credentials = `{
		"username": "test_participant",
		"password": "fmt12345"
	}`
	participantUser, _ := auth.LoginUser(credentials)
	suite.participantToken = *participantUser.Tokens.AccessToken
	suite.participantID = participantUser.UserID
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"title": {
				"en": "Testing trip creation"
			},
			"description": {
				"en": "Trip description test number 1"
			}
		}`,
	}

	trip := trips.Trip{}
	response, err := trip.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &trip)
	suite.tripID = trip.ID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
