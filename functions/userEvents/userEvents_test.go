package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	TableName = "UserEventsTest"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	db          *dynamodb.DynamoDB
	userEventID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.UserEventsTable = TableName
	db.CreateTable(TableName, "userEventId", 1, 1)
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewUserEvent() {
	req := events.APIGatewayProxyRequest{
		Body: `{
			"title": "FMT - Testing user event #1",
			"itineraryID": "0000001",
			"tripID": "00000002",
			"languageCode": "pt"
		}`,
	}

	userEvent := resources.UserEvent{}
	response, err := userEvent.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &userEvent)
	suite.userEventID = userEvent.UserEventID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0020GetAllUserEvent() {
	req := events.APIGatewayProxyRequest{}

	userEvent := resources.UserEvent{}
	response, err := userEvent.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0030UpdateUserEvent() {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": suite.userEventID,
		},
		Body: `{
			"title": "FMT - Testing user event #2"
		}`,
	}

	userEvent := resources.UserEvent{}
	response, err := userEvent.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0040DeleteUserEvent() {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": suite.userEventID,
		},
		Body: `{
			"title": "FMT - Testing user event #2",
		}`,
	}

	userEvent := resources.UserEvent{}
	response, err := userEvent.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
