package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/resources/auth"
	"github.com/feedmytrip/api/resources/highlights"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	token       string
	HighlightID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	user, _ := auth.LoginUser(credentials)
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

	highlight := highlights.Highlight{}
	response, err := highlight.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &highlight)
	suite.HighlightID = highlight.ID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0020GetAllHighlights() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
	}

	highlight := highlights.Highlight{}
	response, err := highlight.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0030UpdateHighlight() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"filter": "rome"
		}`,
		PathParameters: map[string]string{
			"id": suite.HighlightID,
		},
	}

	highlight := highlights.Highlight{}
	response, err := highlight.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0040DeleteHighlight() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"filter": "rome"
		}`,
		PathParameters: map[string]string{
			"id": suite.HighlightID,
		},
	}

	highlight := highlights.Highlight{}
	response, err := highlight.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
