package main

// Basic imports
import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources/auth"
	fmt "github.com/feedmytrip/api/resources/events"
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
	inviteID          string
	itineraryEventID  string
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
		"password": "fmt2018"
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
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
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

func (suite *FeedMyTripAPITestSuite) Test0300SaveNewInvite() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"email": "teste@teste.com"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	invite := trips.Invite{}
	response, err := invite.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &invite)
	suite.inviteID = invite.ID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0310SaveNewInvalidEmailInvite() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"email": "invalidemail"
		}`,
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	invite := trips.Invite{}
	response, err := invite.SaveNew(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0320GetTripInvites() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id": suite.tripID,
		},
	}

	invite := trips.Invite{}
	response, err := invite.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0400SaveNewItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"title": {
				"en": "Testing itinerary event creation"
			},
			"description": {
				"en": "Testing itinerary description fields."
			}
		}`,
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
		},
	}

	event := trips.ItineraryEvent{}
	response, err := event.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &event)
	suite.itineraryEventID = event.ID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0410AddGlobalItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"title": {
				"en": "Testing event creation"
			},
			"description": {
				"en": "testing description fields."
			}
		}`,
	}

	globalEvent := fmt.Event{}
	response, err := globalEvent.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &globalEvent)

	req = events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id":              suite.tripID,
			"itinerary_id":    suite.itineraryID,
			"global_event_id": globalEvent.ID,
		},
		Body: `{
			"begin_offset": 0
		}`,
	}

	event := trips.ItineraryEvent{}
	response, err = event.Add(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0420GetItineraryEvents() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
		},
	}

	event := trips.ItineraryEvent{}
	response, err := event.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0430UpdateItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		Body: `{
			"title.en": "Testing event update",
			"duration": 10800
		}`,
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
			"event_id":     suite.itineraryEventID,
		},
	}

	event := trips.ItineraryEvent{}
	response, err := event.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0993DeleteItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id":           suite.tripID,
			"itinerary_id": suite.itineraryID,
			"event_id":     suite.itineraryEventID,
		},
	}

	event := trips.ItineraryEvent{}
	response, err := event.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0994ForbiddenDeleteTripInvite() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.participantToken,
		},
		PathParameters: map[string]string{
			"id":        suite.tripID,
			"invite_id": suite.inviteID,
		},
	}

	invite := trips.Invite{}
	response, err := invite.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusForbidden, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0995DeleteTripInvite() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.adminToken,
		},
		PathParameters: map[string]string{
			"id":        suite.tripID,
			"invite_id": suite.inviteID,
		},
	}

	invite := trips.Invite{}
	response, err := invite.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
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
