package events

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

//Event represents an event on the database
type Event struct {
	ID                  string             `json:"id" db:"id" lock:"true"`
	Active              bool               `json:"active" db:"active"`
	Title               shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = event.id and title.field = 'title'" embedded:"true" persist:"true"`
	Description         shared.Translation `json:"description" table:"translation" alias:"description" on:"description.parent_id = event.id and title.field = 'description'" embedded:"true" persist:"true"`
	MainCategoryID      string             `json:"main_category_id" db:"main_category_id"`
	MainCategory        shared.Translation `json:"main_category" table:"translation" alias:"main_category" on:"main_category.parent_id = event.main_category_id and main_category.field = 'title'" embedded:"true"`
	SecondaryCategoryID string             `json:"secondary_category_id" db:"secondary_category_id"`
	SecondaryCategory   shared.Translation `json:"secondary_category" table:"translation" alias:"secondary_category" on:"secondary_category.parent_id = event.secondary_category_id and secondary_category.field = 'title'" embedded:"true"`
	CountryID           string             `json:"country_id" db:"country_id"`
	Country             shared.Translation `json:"country" table:"translation" alias:"country" on:"country.parent_id = event.country_id and country.field = 'title'" embedded:"true"`
	RegionID            string             `json:"region_id" db:"region_id"`
	Region              shared.Translation `json:"region" table:"translation" alias:"region" on:"region.parent_id = event.region_id and region.field = 'title'" embedded:"true"`
	CityID              string             `json:"city_id" db:"city_id"`
	City                shared.Translation `json:"city" table:"translation" alias:"city" on:"city.parent_id = event.city_id and city.field = 'title'" embedded:"true"`
	Address             string             `json:"address" db:"address"`
	CreatedBy           string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate         time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy           string             `json:"updated_by" db:"updated_by"`
	UpdatedDate         time.Time          `json:"updated_date" db:"updated_date"`
	CreatedUser         shared.User        `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = event.created_by" embedded:"true"`
	UpdatedUser         shared.User        `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = event.updated_by" embedded:"true"`
}

//Get return an event
func (e *Event) Get(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.QueryOne(session, db.TableEvent, request.PathParameters["id"], Event{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//GetAll returns all events available in the database
func (e *Event) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.Select(session, db.TableEvent, request.QueryStringParameters, Event{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new event
func (e *Event) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), e)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if e.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("invalid request empty title"))
	}

	e.ID = uuid.New().String()
	e.Active = true
	e.Title.ID = uuid.New().String()
	e.Title.Table = db.TableEvent
	e.Title.Field = "title"
	e.Title.ParentID = e.ID
	e.Description.ID = uuid.New().String()
	e.Description.Table = db.TableEvent
	e.Description.Field = "description"
	e.Description.ParentID = e.ID
	e.CreatedBy = tokenUser.UserID
	e.CreatedDate = time.Now()
	e.UpdatedBy = tokenUser.UserID
	e.UpdatedDate = time.Now()

	e.Title.Translate()

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

	err = db.Insert(tx, db.TableEvent, *e)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableEvent, e.ID, Event{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Update change event attributes in the database
func (e *Event) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	if field, ok := request.QueryStringParameters["translate"]; ok {
		if field != "title" && field != "description" {
			return common.APIError(http.StatusBadRequest, errors.New("invalid translation field"))
		}
		translation := shared.Translation{}
		if val, ok := jsonMap[field+".en"]; ok {
			translation.EN = val.(string)
		} else if val, ok := jsonMap[field+".pt"]; ok {
			translation.PT = val.(string)
		} else if val, ok := jsonMap[field+".es"]; ok {
			translation.ES = val.(string)
		}
		if !translation.IsEmpty() {
			translation.Translate()
			jsonMap[field+".en"] = translation.EN
			jsonMap[field+".es"] = translation.ES
			jsonMap[field+".pt"] = translation.PT
		}
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

	err = db.Update(tx, db.TableEvent, request.PathParameters["id"], *e, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableEvent, request.PathParameters["id"], Event{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete removes event from the database
func (e *Event) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	err = db.Delete(session, db.TableEvent, request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
