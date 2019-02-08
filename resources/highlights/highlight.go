package highlights

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources/shared"
	"github.com/google/uuid"
)

//Highlight represents a group of trips and events to emphasize
type Highlight struct {
	ID           string             `json:"id" db:"id" lock:"true"`
	Active       bool               `json:"active" db:"active"`
	Title        shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = highlight.id and title.field = 'title'" embedded:"true" persist:"true"`
	Description  shared.Translation `json:"description" table:"translation" alias:"description" on:"description.parent_id = highlight.id and description.field = 'description'" embedded:"true" persist:"true"`
	ScheduleDate time.Time          `json:"schedule_date" db:"schedule_date"`
	Filter       string             `json:"filter" db:"filter"`
	CountryID    string             `json:"country_id" db:"country_id"`
	Country      shared.Translation `json:"country" table:"translation" alias:"country" on:"country.parent_id = highlight.country_id and country.field = 'title'" embedded:"true"`
	RegionID     string             `json:"region_id" db:"region_id"`
	Region       shared.Translation `json:"region" table:"translation" alias:"region" on:"region.parent_id = highlight.region_id and region.field = 'title'" embedded:"true"`
	CityID       string             `json:"city_id" db:"city_id"`
	City         shared.Translation `json:"city" table:"translation" alias:"city" on:"city.parent_id = highlight.city_id and city.field = 'title'" embedded:"true"`
	Trips        string             `json:"trips" db:"trips"`
	Events       string             `json:"events" db:"events"`
	CreatedBy    string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate  time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy    string             `json:"updated_by" db:"updated_by"`
	UpdatedDate  time.Time          `json:"updated_date" db:"updated_date"`
	CreatedUser  shared.User        `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = highlight.created_by" embedded:"true"`
	UpdatedUser  shared.User        `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = highlight.updated_by" embedded:"true"`
}

//Get return a highlight
func (h *Highlight) Get(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.QueryOne(session, db.TableHighlight, request.PathParameters["id"], Highlight{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//GetAll returns all highlights available in the database
func (h *Highlight) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.Select(session, db.TableHighlight, request.QueryStringParameters, Highlight{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new highlight
func (h *Highlight) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), h)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if h.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("invalid request empty title"))
	}

	h.ID = uuid.New().String()
	h.Active = true
	h.Title.ID = uuid.New().String()
	h.Title.Table = db.TableHighlight
	h.Title.Field = "title"
	h.Title.ParentID = h.ID
	h.Description.ID = uuid.New().String()
	h.Description.Table = db.TableHighlight
	h.Description.Field = "description"
	h.Description.ParentID = h.ID
	h.CreatedBy = tokenUser.UserID
	h.CreatedDate = time.Now()
	h.UpdatedBy = tokenUser.UserID
	h.UpdatedDate = time.Now()

	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Insert(tx, db.TableHighlight, *h)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableHighlight, h.ID, Highlight{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Update change highlight attributes in the database
func (h *Highlight) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Update(tx, db.TableHighlight, request.PathParameters["id"], *h, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableHighlight, request.PathParameters["id"], Highlight{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete removes highlight from the database
func (h *Highlight) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	//TODO: Delete all highlight images and then delete highlight folder on AWS S3

	err = db.Delete(session, db.TableHighlight, request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
