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
	adminToken        string
	participantToken  string
	participantUserID string
	participantID     string
	itineraryID       string
	tripID            string
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
	suite.participantUserID = participantUser.UserID
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

func (suite *FeedMyTripAPITestSuite) Test0020GetAllTrips() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
	}

	trip := trips.Trip{}
	response, err := trip.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0030UpdateTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"description.en": "Edited description using patch method"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
		QueryStringParameters: map[string]string{
			"translate": "description",
		},
	}

	trip := trips.Trip{}
	response, err := trip.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0100ForbiddenGetTripParticipants() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.participantToken,
		},
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	participant := trips.Participant{}
	response, err := participant.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusForbidden, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0110SaveNewParticipant() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"user_id": "` + suite.participantUserID + `",
			"role": "` + trips.ParticipantAdminRole + `"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	participant := trips.Participant{}
	response, err := participant.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &participant)
	suite.participantID = participant.ID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0120GetTripParticipants() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	participant := trips.Participant{}
	response, err := participant.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0130UpdateParticipant() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"role": "` + trips.ParticipantViewerRole + `"
		}`,
		PathParameters: map[string]string{
			"id":             suite.tripID,
			"participant_id": suite.participantID,
		},
	}

	participant := trips.Participant{}
	response, err := participant.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0200ForbiddenSaveNewItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.participantToken,
		},
		Body: `{
			"title.pt": "Novo Roteiro"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.SaveNew(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusForbidden, response.StatusCode, response.Body)
}
func (suite *FeedMyTripAPITestSuite) Test0210SaveNewItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"title.pt": "Novo Roteiro"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &itinerary)
	suite.itineraryID = itinerary.ID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0220GetTripItineraries() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0230UpdateItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"title.pt": "Novo roteiro atualizado"
		}`,
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0240ForbiddenUpdateItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.participantToken,
		},
		Body: `{
			"title.pt": "Novo roteiro atualizado"
		}`,
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusForbidden, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0996ForbiddenDeleteItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.participantToken,
		},
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusForbidden, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0997DeleteItinerary() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
		},
	}

	itinerary := trips.Itinerary{}
	response, err := itinerary.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0998DeleteParticipant() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id":             suite.tripID,
			"participant_id": suite.participantID,
		},
	}

	participant := trips.Participant{}
	response, err := participant.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0999InvalidDeleteTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.participantToken,
		},
		PathParameters: map[string]string{
			"id":             suite.tripID,
			"participant_id": suite.participantID,
		},
	}

	trip := trips.Trip{}
	response, err := trip.Delete(req)
	json.Unmarshal([]byte(response.Body), &trip)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusForbidden, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test1000DeleteTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	trip := trips.Trip{}
	response, err := trip.Delete(req)
	json.Unmarshal([]byte(response.Body), &trip)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
