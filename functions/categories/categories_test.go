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
	TableName = "CategoriesTest"
)

type FeedMyTripAPITestSuite struct {
	suite.Suite
	db         *dynamodb.DynamoDB
	categoryID string
}

func (suite *FeedMyTripAPITestSuite) SetupTest() {
	common.CategoriesTable = TableName
	db.CreateTable(TableName, "categoryId", 1, 1)
}

func (suite *FeedMyTripAPITestSuite) Test0010SaveNewCategory() {
	req := events.APIGatewayProxyRequest{
		Body: `{
			"title": {
				"en": "Transports",
				"pt": "Transportes",
				"es": "Transportes"
			}
		}`,
	}

	category := resources.Category{}
	response, err := category.SaveNew(req)
	json.Unmarshal([]byte(response.Body), &category)
	suite.categoryID = category.CategoryID

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0020GetAllCategories() {
	req := events.APIGatewayProxyRequest{}

	category := resources.Category{}
	response, err := category.GetAll(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0021GetAllActiveCategories() {
	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"state": "active",
		},
	}

	category := resources.Category{}
	response, err := category.GetAll(req)
	list := []resources.Category{}
	json.Unmarshal([]byte(response.Body), &list)
	active := false
	if len(list) > 0 {
		active = list[0].Active
	}

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
	assert.Equal(suite.T(), active, true)
}

func (suite *FeedMyTripAPITestSuite) Test0030UpdateCategory() {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": suite.categoryID,
		},
		Body: `{
			"active": false,
			"title.en": "Lodge"
		}`,
	}

	category := resources.Category{}
	response, err := category.Update(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *FeedMyTripAPITestSuite) Test0040DeleteCategory() {
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": suite.categoryID,
		},
	}

	category := resources.Category{}
	response, err := category.Delete(req)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func TestFeedMyTripAPITestSuite(t *testing.T) {
	suite.Run(t, new(FeedMyTripAPITestSuite))
}
