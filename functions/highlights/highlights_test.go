package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	TableName = "HighlightsTest"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	token       string
	HighlightID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.HighlightsTable = TableName
	db.CreateTable(TableName, "highlightId", 1, 1)

	credentials := `{
		"username": "test",
		"password": "test12345"
	}`
	user, _ := resources.LoginUser(credentials)
	suite.token = *user.Tokens.AccessToken
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewHighlight() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"title": {
				"pt": "Novo destaque",
				"en": "New highlight",
				"es": "Nuevo punto culminante"
			}
		}`,
	}

	highlight := resources.Highlight{}
	response, err := highlight.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &highlight)
	suite.HighlightID = highlight.HighlightID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
