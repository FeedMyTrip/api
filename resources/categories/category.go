package categories

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/google/uuid"

	"github.com/feedmytrip/api/resources/shared"
)

//Category represents a category in the system
type Category struct {
	ID             string             `json:"id" db:"id" lock:"true"`
	ParentID       string             `json:"parent_id" db:"parent_id"`
	Active         bool               `json:"active" db:"active"`
	ParentCategory shared.Translation `json:"parent_category" table:"translation" alias:"parent_category" on:"parent_category.parent_id = category.parent_id and parent_category.field = 'title'" embedded:"true"`
	Title          shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = category.id and title.field = 'title'" embedded:"true" persist:"true"`
	CreatedBy      string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate    time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy      string             `json:"updated_by" db:"updated_by"`
	UpdatedDate    time.Time          `json:"updated_date" db:"updated_date"`
	CreatedUser    shared.User        `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = category.created_by" embedded:"true"`
	UpdatedUser    shared.User        `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = category.updated_by" embedded:"true"`
}

//GetAll returns all categories available in the database
func (c *Category) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.Select(session, db.TableCategory, request.QueryStringParameters, Category{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new category
func (c *Category) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), c)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if c.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("empty title"))
	}

	c.ID = uuid.New().String()
	c.Active = true
	c.Title.ID = uuid.New().String()
	c.Title.ParentID = c.ID
	c.Title.Table = db.TableCategory
	c.Title.Field = "title"
	c.CreatedBy = tokenUser.UserID
	c.CreatedDate = time.Now()
	c.UpdatedBy = tokenUser.UserID
	c.UpdatedDate = time.Now()

	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()
	defer session.Close()
	defer conn.Close()

	err = db.Insert(tx, db.TableCategory, *c)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableCategory, c.ID, Category{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusCreated)
}

//Update change categories attributes in the database
func (c *Category) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap["updated_by"] = tokenUser.UserID
	jsonMap["updated_date"] = time.Now()

	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()
	defer session.Close()
	defer conn.Close()

	err = db.Update(tx, db.TableCategory, request.PathParameters["id"], *c, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableCategory, request.PathParameters["id"], Category{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete removes categories from the database
func (c *Category) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	err = db.Delete(session, db.TableCategory, request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
