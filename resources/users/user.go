package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources/shared"
)

//User represents a user in the system
type User struct {
	ID              string             `json:"id" db:"id" lock:"true"`
	Active          bool               `json:"active" db:"active" lock:"true"`
	FirstName       string             `json:"first_name" db:"first_name" lock:"true" filter:"true"`
	LastName        string             `json:"last_name" db:"last_name" lock:"true" filter:"true"`
	Group           string             `json:"group" db:"group" lock:"true" filter:"true"`
	Username        string             `json:"username" db:"username" lock:"true" filter:"true"`
	Email           string             `json:"email" db:"email" lock:"true" filter:"true"`
	LanguageCode    string             `json:"language_code" db:"language_code" lock:"true"`
	PrincipalTripID string             `json:"principal_trip_id" db:"principal_trip_id"`
	ImagePath       string             `json:"image_path" db:"image_path"`
	CountryID       string             `json:"country_id" db:"country_id"`
	Country         shared.Translation `json:"country" table:"translation" alias:"country" on:"country.parent_id = user.country_id and country.field = 'title'" embedded:"true"`
	RegionID        string             `json:"region_id" db:"region_id"`
	Region          shared.Translation `json:"region" table:"translation" alias:"region" on:"region.parent_id = user.region_id and region.field = 'title'" embedded:"true"`
	CityID          string             `json:"city_id" db:"city_id"`
	City            shared.Translation `json:"city" table:"translation" alias:"city" on:"city.parent_id = user.city_id and city.field = 'title'" embedded:"true"`
	AboutMe         string             `json:"about_me" db:"about_me"`
	CreatedBy       string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate     time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy       string             `json:"updated_by" db:"updated_by"`
	UpdatedDate     time.Time          `json:"updated_date" db:"updated_date"`
	CreatedUser     shared.User        `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = user.created_by" embedded:"true"`
	UpdatedUser     shared.User        `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = user.updated_by" embedded:"true"`
}

//GetAll returns all users available in the database
func (u *User) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	result, err := db.Select(session, db.TableUser, request.QueryStringParameters, User{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new user
func (u *User) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), u)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	u.Active = true
	u.CreatedBy = tokenUser.UserID
	u.CreatedDate = time.Now()
	u.UpdatedBy = tokenUser.UserID
	u.UpdatedDate = time.Now()

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

	err = db.Insert(tx, db.TableUser, *u)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(u, http.StatusCreated)
}

//Update change user attributes in the database
func (u *User) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	err = db.Update(tx, db.TableUser, request.PathParameters["id"], *u, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableUser, request.PathParameters["id"], User{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete removes user from the database
func (u *User) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	err = db.Delete(session, db.TableUser, request.PathParameters["id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
