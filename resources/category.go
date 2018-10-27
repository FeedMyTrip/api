package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/google/uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

//Category represents a way to index the events on the system
type Category struct {
	CategoryID     string      `json:"categoryId" validate:"required"`
	MainCategoryID string      `json:"mainCategoryId"`
	Title          Translation `json:"title" validate:"required"`
	Active         bool        `json:"active"`
	Audit          *Audit      `json:"audit"`
}

//GetAll returns all categories available in the database
func (c *Category) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterExpression, filterValues := common.ParseRequestFilters(request)
	result, err := db.Scan(common.CategoriesTable, filterExpression, filterValues)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	list := []Category{}
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &list)
	return common.APIResponse(list, http.StatusOK)
}

//SaveNew creates a new category
func (c *Category) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), c)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	c.CategoryID = uuid.New().String()
	c.Audit = NewAudit(common.GetTokenUser(request).UserID)
	c.Active = true

	validate := validator.New()
	err = validate.Struct(c)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	err = db.PutItem(c, common.CategoriesTable)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(c, http.StatusCreated)
}

//Update modify category attributes
func (c *Category) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO check if body is valid
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request)
	jsonMap["audit.updatedDate"] = time.Now()

	result, err := db.UpdateItem(common.CategoriesTable, "categoryId", request.PathParameters["id"], jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	dynamodbattribute.UnmarshalMap(result.Attributes, c)
	return common.APIResponse(c, http.StatusOK)
}

//Delete removes a new category from the database
func (c *Category) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	//TODO implement mark to delete
	//TODO verify if there is any event with this category
	err := db.DeleteItem(common.CategoriesTable, "categoryId", request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}
