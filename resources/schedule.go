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

//Schedule represents all schedules and availabilities for an event
type Schedule struct {
	ScheduleID  string    `json:"scheduleId"`
	Annualy     bool      `json:"annualy"`
	FixedDate   bool      `json:"fixedDate"`
	FixedPeriod bool      `json:"fixedPeriod"`
	Closed      bool      `json:"closed"`
	StartDate   time.Time `json:"startDate" validate:"required"`
	EndDate     time.Time `json:"endDate" validate:"required"`
	WeekDays    string    `json:"weekDays" validate:"required"`
	Audit       *Audit    `json:"audit"`
}

//ScheduleResponse represents the response with the event id and the schedule object
type ScheduleResponse struct {
	EventID  string    `json:"eventId"`
	Schedule *Schedule `json:"schedule"`
}

//SaveNew creates a new schedule for an event
func (s *Schedule) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	err := json.Unmarshal([]byte(request.Body), s)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	s.ScheduleID = uuid.New().String()
	s.Audit = NewAudit(common.GetTokenUser(request).UserID)

	validate := validator.New()
	err = validate.Struct(s)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap := make(map[string]interface{})
	schedules := []Schedule{}
	schedules = append(schedules, *s)
	jsonMap[":schedules"] = schedules

	_, err = db.PutListItem(common.EventsTable, "eventId", request.PathParameters["id"], "schedules", jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateEventAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	sr := ScheduleResponse{}
	sr.EventID = request.PathParameters["id"]
	sr.Schedule = s

	return common.APIResponse(sr, http.StatusCreated)
}

//Update saves schedule modifications to the database
func (s *Schedule) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	//TODO change to user id that executes the action
	jsonMap["audit.updatedBy"] = common.GetTokenUser(request).UserID
	jsonMap["audit.updatedDate"] = time.Now()

	event := Event{}
	event.Load(request)
	index := getScheduleIndex(event.Schedules, request.PathParameters["scheduleId"])
	if index == -1 {
		return common.APIError(http.StatusNotFound, errors.New("schedule not found"))
	}

	result, err := db.UpdateListItem(common.EventsTable, "eventId", event.EventID, "schedules", index, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = dynamodbattribute.UnmarshalMap(result.Attributes, &event)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = UpdateEventAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	sr := ScheduleResponse{}
	sr.EventID = request.PathParameters["id"]
	sr.Schedule = &event.Schedules[index]
	return common.APIResponse(sr, http.StatusOK)
}

//Delete removes an event schedule
func (s *Schedule) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !common.IsTokenUserAdmin(request) {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	event := Event{}
	event.Load(request)

	index := getScheduleIndex(event.Schedules, request.PathParameters["scheduleId"])
	if index == -1 {
		return common.APIError(http.StatusNotFound, errors.New("schedule not found"))
	}

	err := db.DeleteListItem(common.EventsTable, "eventId", request.PathParameters["id"], "schedules", index)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func getScheduleIndex(schedules []Schedule, scheduleID string) int {
	index := 0
	for _, es := range schedules {
		if es.ScheduleID == scheduleID {
			return index
		}
		index++
	}
	return -1
}
