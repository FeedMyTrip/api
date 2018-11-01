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
	TableName = "GeonamesTest"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	db        *dynamodb.DynamoDB
	token     string
	countryID string
	cityID    string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.GeonamesTable = TableName
	db.CreateTable(TableName, "geonameId", 1, 1)

	credentials := `{
		"username": "test_admin",
		"password": "fmt12345"
	}`
	user, _ := resources.LoginUser(credentials)
	suite.token = *user.Tokens.AccessToken
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewCountry() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"translation": {
				"pt": "Brasil",
				"en": "Brazil",
				"es": "Brasil"
			}
		}`,
	}

	geoname := resources.Geoname{}
	response, err := geoname.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &geoname)
	suite.countryID = geoname.GeonameID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test0020SaveNewCity() {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		Body: `{
			"countryId": "` + suite.countryID + `",
			"translation": {
				"pt": "Rio de Janeiro",
				"en": "Rio de Janeiro",
				"es": "Rio de Janeiro"
			}
		}`,
	}

	geoname := resources.Geoname{}
	response, err := geoname.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &geoname)
	suite.cityID = geoname.GeonameID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode, response.Body)
}

func (suite *FeedMyTripAPITestSuite) Test1000Delete() {
	reqCountry := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id": suite.countryID,
		},
	}

	reqCity := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization": suite.token,
		},
		PathParameters: map[string]string{
			"id": suite.cityID,
		},
	}

	geoname := resources.Geoname{}
	response, err := geoname.Delete(reqCountry)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)

	response, err = geoname.Delete(reqCity)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode, response.Body)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
