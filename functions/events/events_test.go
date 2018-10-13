package main

import (
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
	TableName = "EventsTest"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	db      *dynamodb.DynamoDB
	eventID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.TripsTable = TableName
	db.CreateTable(TableName, "eventId", 1, 1)
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewTrip() {
	req := events.APIGatewayProxyRequest{
		Body: `{
			"title": "FMT - Testing suite #1",
			"description": "Loren ipsum ea est atqui iisque placerat, est nobis videre."
		}`,
	}

	event := resources.Event{}
	response, err := event.SaveNew(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
