package main

// Basic imports
import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/feedmytrip/api/common"

	"github.com/feedmytrip/api/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	TableName = "TripsTest"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	db                   *dynamodb.DynamoDB
	token                string
	tripID               string
	participantID        string
	participantOwnerID   string
	inviteID             string
	itineraryID          string
	principalItineraryID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.TripsTable = TableName
	db.CreateTable(TableName, "tripId", 1, 1)

	credentials := `{
		"username": "test",
		"password": "test12345"
	}`
	user, _ := resources.LoginUser(credentials)
	suite.token = *user.Tokens.AccessToken
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": "FMT - Testing suite #1",
			"description": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
		}`,
	}

	trip := resources.Trip{}
	response, err := trip.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &trip)
	suite.tripID = trip.TripID
	suite.participantOwnerID = trip.Participants[0].ParticipantID
	suite.principalItineraryID = trip.ItineraryID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0020SaveNewTripEmptyTitleFail() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"description": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
		}`,
	}

	trip := resources.Trip{}
	response, err := trip.SaveNew(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0030GetAllTrips() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
	}

	trip := resources.Trip{}
	response, err := trip.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0040UpdateTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"description": "Edited description using patch: Loren ipsum ea est atqui iisque placerat, est nobis videre."
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	trip := resources.Trip{}
	response, err := trip.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0050SaveNewParticipant() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"userId": "000005",
			"userRole": "Viewer"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	participant := resources.Participant{}
	response, err := participant.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &participant)
	suite.participantID = participant.ParticipantID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0060UpdateParticipant() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"userRole": "Editor"
		}`,
		PathParameters: map[string]string{
			"id":            suite.tripID,
			"participantId": suite.participantID,
		},
	}

	participant := resources.Participant{}
	response, err := participant.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0070DeleteParticipantOwnerFail() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id":            suite.tripID,
			"participantId": suite.participantOwnerID,
		},
	}

	participant := resources.Participant{}
	response, err := participant.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0080DeleteParticipant() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id":            suite.tripID,
			"participantId": suite.participantID,
		},
	}

	participant := resources.Participant{}
	response, err := participant.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0090SaveNewInvite() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"email": "teste@email.com"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	invite := resources.Invite{}
	response, err := invite.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &invite)
	suite.inviteID = invite.InviteID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0100DeleteInvite() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id":       suite.tripID,
			"inviteId": suite.inviteID,
		},
	}

	invite := resources.Invite{}
	response, err := invite.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0110SaveNewItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": "Europe 2019 - France and Italy",
			"userId": "000001",
			"startDate": "2018-10-12T02:46:13.164772488Z",
			"endDate": "2018-10-27T02:46:13.164772488Z"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	itinerary := resources.Itinerary{}
	response, err := itinerary.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &itinerary)
	suite.itineraryID = itinerary.ItineraryID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0120UpdateItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"startDate": "2018-11-12T02:46:13.164772488Z",
			"endDate": "2018-11-27T02:46:13.164772488Z"
		}`,
		PathParameters: map[string]string{
			"id":          suite.tripID,
			"itineraryId": suite.itineraryID,
		},
	}

	itinerary := resources.Itinerary{}
	response, err := itinerary.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0130DeletePrincipalItineraryFail() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"startDate": "2018-11-12T02:46:13.164772488Z",
			"endDate": "2018-11-27T02:46:13.164772488Z"
		}`,
		PathParameters: map[string]string{
			"id":          suite.tripID,
			"itineraryId": suite.principalItineraryID,
		},
	}

	itinerary := resources.Itinerary{}
	response, err := itinerary.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0130DeleteItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"startDate": "2018-11-12T02:46:13.164772488Z",
			"endDate": "2018-11-27T02:46:13.164772488Z"
		}`,
		PathParameters: map[string]string{
			"id":          suite.tripID,
			"itineraryId": suite.itineraryID,
		},
	}

	itinerary := resources.Itinerary{}
	response, err := itinerary.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test1000DeleteTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	trip := resources.Trip{}
	response, err := trip.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
