package trips

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gocraft/dbr"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/google/uuid"
)

const (
	//ParticipantOwnerRole defines a participant as a trip owner
	ParticipantOwnerRole = "owner"
	//ParticipantAdminRole defines a participant as a trip admin
	ParticipantAdminRole = "admin"
	//ParticipantEditorRole defines a participant as a trip editor
	ParticipantEditorRole = "editor"
	//ParticipantViewerRole defines a participant as a trip viewer
	ParticipantViewerRole = "viewer"
)

// Participant represents a user that is participating in the trip
type Participant struct {
	ID          string    `json:"id" db:"id" lock:"true"`
	TripID      string    `json:"trip_id" db:"trip_id" lock:"true"`
	UserID      string    `json:"user_id" db:"user_id" lock:"true"`
	Role        string    `json:"role" db:"role"`
	CreatedBy   string    `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	UpdatedDate time.Time `json:"updated_date" db:"updated_date"`
}

//GetAll returns all participant a from the trip
func (p *Participant) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		filter := dbr.And(
			dbr.Eq("trip_id", request.PathParameters["id"]),
			dbr.Eq("user_id", tokenUser.UserID),
		)
		total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("only trip participant can access this resource"))
		}
	}

	if request.QueryStringParameters == nil {
		request.QueryStringParameters = map[string]string{
			"trip_id": request.PathParameters["id"],
		}
	} else {
		request.QueryStringParameters["trip_id"] = request.PathParameters["id"]
	}

	result, err := db.Select(session, db.TableTripParticipant, request.QueryStringParameters, Participant{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew add a new participant to the trip
func (p *Participant) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		filter := dbr.And(
			dbr.Eq("trip_id", request.PathParameters["id"]),
			dbr.Eq("user_id", tokenUser.UserID),
			dbr.Or(
				dbr.Eq("role", ParticipantOwnerRole),
				dbr.Eq("role", ParticipantAdminRole),
			),
		)
		total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("only trip owner or admins can make changes"))
		}
	}

	err = json.Unmarshal([]byte(request.Body), p)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	p.ID = uuid.New().String()
	p.TripID = request.PathParameters["id"]
	p.CreatedBy = tokenUser.UserID
	p.CreatedDate = time.Now()
	p.UpdatedBy = tokenUser.UserID
	p.UpdatedDate = time.Now()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Insert(tx, db.TableTripParticipant, *p)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	return common.APIResponse(p, http.StatusCreated)
}

//Update change participant attributes
func (p *Participant) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	filter := dbr.And(
		dbr.Eq("id", request.PathParameters["participant_id"]),
		dbr.Eq("trip_id", request.PathParameters["id"]),
		dbr.Neq("role", ParticipantOwnerRole),
	)
	total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	if total <= 0 {
		return common.APIError(http.StatusForbidden, errors.New("trip owner can't be updated"))
	}

	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		filter := dbr.And(
			dbr.Eq("trip_id", request.PathParameters["id"]),
			dbr.Eq("user_id", tokenUser.UserID),
			dbr.Or(
				dbr.Eq("role", ParticipantOwnerRole),
				dbr.Eq("role", ParticipantAdminRole),
			),
		)
		total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("only trip admin can update participants"))
		}
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	jsonMap["updated_by"] = tokenUser.UserID
	jsonMap["updated_date"] = time.Now()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Update(tx, db.TableTripParticipant, request.PathParameters["participant_id"], *p, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableTripParticipant, request.PathParameters["participant_id"], Participant{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete remove participant
func (p *Participant) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	filter := dbr.And(
		dbr.Eq("id", request.PathParameters["participant_id"]),
		dbr.Eq("trip_id", request.PathParameters["id"]),
		dbr.Neq("role", ParticipantOwnerRole),
	)
	total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	if total <= 0 {
		return common.APIError(http.StatusForbidden, errors.New("trip owner can't be deleted"))
	}

	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		filter := dbr.And(
			dbr.Eq("trip_id", request.PathParameters["id"]),
			dbr.Eq("user_id", tokenUser.UserID),
			dbr.Or(
				dbr.Eq("role", ParticipantOwnerRole),
				dbr.Eq("role", ParticipantAdminRole),
			),
		)
		total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("only trip participant can access this resource"))
		}
	}

	err = db.Delete(session, db.TableTripParticipant, request.PathParameters["participant_id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
