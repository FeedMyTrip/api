package locations

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources/shared"
	"github.com/google/uuid"
)

//Location represents a country, a region or a city in the system
type Location struct {
	ID        string             `json:"id" db:"id" lock:"true"`
	CountryID string             `json:"country_id" db:"country_id"`
	RegionID  string             `json:"region_id" db:"region_id"`
	Title     shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = location.id and title.field = 'title'" embedded:"true" persist:"true"`
}

//SaveNew creates a new country or city
func (l *Location) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), l)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if l.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("empty title"))
	}

	l.ID = uuid.New().String()
	l.Title.ID = uuid.New().String()
	l.Title.Table = db.TableLocation
	l.Title.Field = "title"
	l.Title.ParentID = l.ID

	l.Title.Translate()

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

	err = db.Insert(tx, db.TableLocation, *l)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(l, http.StatusCreated)
}

//GetAll returns a list of locations
func (l *Location) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.Select(session, db.TableLocation, request.QueryStringParameters, Location{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Update change location attributes
func (l *Location) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

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

	err = db.Update(tx, db.TableLocation, request.PathParameters["id"], *l, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(l, http.StatusOK)
}

//Delete remove location from the database
func (l *Location) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	err = db.Delete(session, db.TableLocation, request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
