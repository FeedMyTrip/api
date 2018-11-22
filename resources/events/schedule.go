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

//Schedule represents all schedules and availabilities for an event
type Schedule struct {
	ID          string      `json:"id" db:"id" lock:"true"`
	EventID     string      `json:"event_id" db:"event_id" lock:"true"`
	Annually    bool        `json:"annually" db:"annually"`
	FixedDate   bool        `json:"fixed_date" db:"fixed_date"`
	FixedPeriod bool        `json:"fixed_period" db:"fixed_period"`
	Closed      bool        `json:"closed" db:"closed"`
	StartDate   time.Time   `json:"start_date" db:"start_date"`
	EndDate     time.Time   `json:"end_date" db:"end_date"`
	WeekDays    string      `json:"week_days" db:"week_days"`
	CreatedBy   string      `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time   `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy   string      `json:"updated_by" db:"updated_by"`
	UpdatedDate time.Time   `json:"updated_date" db:"updated_date"`
	CreatedUser shared.User `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = event_schedule.created_by" embedded:"true"`
	UpdatedUser shared.User `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = event_schedule.updated_by" embedded:"true"`
}

//GetAll returns all schedules for an event available in the database
func (s *Schedule) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()
	defer conn.Close()

	if request.QueryStringParameters == nil {
		request.QueryStringParameters = map[string]string{
			"event_id": request.PathParameters["id"],
		}
	} else {
		request.QueryStringParameters["event_id"] = request.PathParameters["id"]
	}

	result, err := db.Select(session, db.TableEventSchedule, request.QueryStringParameters, Schedule{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new schedule for the event
func (s *Schedule) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	err := json.Unmarshal([]byte(request.Body), s)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	s.ID = uuid.New().String()
	s.EventID = request.PathParameters["id"]
	s.CreatedBy = tokenUser.UserID
	s.CreatedDate = time.Now()
	s.UpdatedBy = tokenUser.UserID
	s.UpdatedDate = time.Now()

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

	err = db.Insert(tx, db.TableEventSchedule, *s)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(s, http.StatusCreated)
}

//Update change event schedule attributes in the database
func (s *Schedule) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	err = db.Update(tx, db.TableEventSchedule, request.PathParameters["schedule_id"], *s, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(nil, http.StatusOK)
}

//Delete removes event schedule from the database
func (s *Schedule) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	err = db.Delete(session, db.TableEventSchedule, request.PathParameters["schedule_id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
