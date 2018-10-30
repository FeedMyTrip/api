package main

// Basic imports
import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/feedmytrip/api/common"

	"github.com/feedmytrip/api/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	token                string
	loggedUserID         string
	tripID               string
	participantID        string
	participantOwnerID   string
	participantUserId    string
	inviteID             string
	itineraryID          string
	principalItineraryID string
	itineraryEventID     string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.TripsTable = "TripsTest"
	common.UsersTable = "UsersTest"

	db.CreateTable("TripsTest", "tripId", 1, 1)

	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	loggedUser, _ := resources.LoginUser(credentials)
	suite.token = *loggedUser.Tokens.AccessToken
	suite.loggedUserID = loggedUser.UserID

	credentials = `{
		"username": "test_participant",
		"password": "fmt12345"
	}`
	participantUser, _ := resources.LoginUser(credentials)
	suite.participantUserId = participantUser.UserID

	user := resources.User{}
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id": suite.loggedUserID,
		},
	}
	user.GetUserDetails(req)

	req.PathParameters["id"] = suite.participantUserId
	participant := resources.User{}
	participant.GetUserDetails(req)
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewTrip() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": {
				"en": "FMT - Testing suite #1"
			},
			"description": {
				"en": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
			}
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
			"description.pt": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
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
			"description.en": "Edited description using patch: Loren ipsum ea est atqui iisque placerat, est nobis videre."
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
			"userId": "` + suite.participantUserId + `",
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
			"title": {
				"en": "FMT - Testing suite #1"
			},
			"userId": "` + suite.loggedUserID + `",
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

func (suite *FeedMyTripAPITestSuite) Test0130CreateNewItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": {
				"en": "FMT - Testing suite #1"
			},
			"description": {
				"en": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
			}
		}`,
		PathParameters: map[string]string{
			"id":          suite.tripID,
			"itineraryId": suite.itineraryID,
		},
	}

	event := resources.UserEvent{}
	response, err := event.SaveNew(req)

	uer := resources.UserEventResponse{}
	json.Unmarshal([]byte(response.Body), &uer)
	suite.itineraryEventID = uer.Event.UserEventID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0140UpdateItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title.pt": "FMT - Testando #1",
			"description.pt": "Loren ipsum ea est atqui iisque placerat, est nobis videre.",
			"beginOffset": 259200
			}`,
		PathParameters: map[string]string{
			"id":          suite.tripID,
			"itineraryId": suite.itineraryID,
			"eventId":     suite.itineraryEventID,
		},
	}

	event := resources.UserEvent{}
	response, err := event.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0150DeleteItineraryEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id":          suite.tripID,
			"itineraryId": suite.itineraryID,
			"eventId":     suite.itineraryEventID,
		},
	}

	event := resources.UserEvent{}
	response, err := event.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0160AddItineraryGlobalEvent() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": {
				"pt": "FMT - Testing suite #1"
			},
			"description": {
				"pt": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
			}
		}`,
	}

	globalEvent := resources.Event{}
	response, err := globalEvent.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &globalEvent)
	globalEventID := globalEvent.EventID

	req = events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id":            suite.tripID,
			"itineraryId":   suite.itineraryID,
			"globalEventId": globalEventID,
		},
	}

	event := resources.UserEvent{}
	response, err = event.AddGlobal(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0170DeletePrincipalItineraryFail() {
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

func (suite *FeedMyTripAPITestSuite) Test0180DeleteItinerary() {
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
