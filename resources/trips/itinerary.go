package trips

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources/shared"
	"github.com/gocraft/dbr"
	"github.com/google/uuid"
)

// Itinerary represents a way to group trip events
type Itinerary struct {
	ID          string             `json:"id" db:"id" lock:"true"`
	TripID      string             `json:"trip_id" db:"trip_id" lock:"true"`
	Title       shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = trip_itinerary.id and title.field = 'title'" embedded:"true" persist:"true"`
	OwnerID     string             `json:"owner_id" db:"owner_id"`
	StartDate   time.Time          `json:"start_date" db:"start_date"`
	EndDate     time.Time          `json:"end_date" db:"end_date"`
	CreatedBy   string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy   string             `json:"updated_by" db:"updated_by"`
	UpdatedDate time.Time          `json:"updated_date" db:"updated_date"`
}

//GetAll returns all itineraries from the trip
func (i *Itinerary) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	result, err := db.Select(session, db.TableTripItinerary, request.QueryStringParameters, Itinerary{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew add a new itinerary to the trip
func (i *Itinerary) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			dbr.Neq("role", ParticipantViewerRole),
		)
		total, err := db.Validate(session, []string{"count(id) total"}, db.TableTripParticipant, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("viewer participants can't create itineraries"))
		}
	}

	err = json.Unmarshal([]byte(request.Body), i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	i.ID = uuid.New().String()
	i.TripID = request.PathParameters["id"]
	i.Title.ID = uuid.New().String()
	i.Title.ParentID = i.ID
	i.Title.Table = db.TableTripItinerary
	i.Title.Field = "title"
	i.OwnerID = tokenUser.UserID
	i.StartDate = time.Now()
	i.EndDate = time.Now()
	i.CreatedBy = tokenUser.UserID
	i.CreatedDate = time.Now()
	i.UpdatedBy = tokenUser.UserID
	i.UpdatedDate = time.Now()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Insert(tx, db.TableTripItinerary, *i)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	return common.APIResponse(i, http.StatusCreated)
}

//Update change itinerary attributes
func (i *Itinerary) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			dbr.Eq("i.id", request.PathParameters["itinerary_id"]),
			dbr.Eq("p.trip_id", request.PathParameters["id"]),
			dbr.Eq("p.user_id", tokenUser.UserID),
			dbr.Or(
				dbr.Eq("i.created_by", tokenUser.UserID),
				dbr.Eq("p.role", ParticipantOwnerRole),
				dbr.Eq("p.role", ParticipantAdminRole),
			),
		)
		table := db.TableTripParticipant + " p , " + db.TableTripItinerary + " i"
		total, err := db.Validate(session, []string{"count(p.id) total"}, table, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("not allowed to update this itinerary"))
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

	err = db.Update(tx, db.TableTripItinerary, request.PathParameters["itinerary_id"], *i, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableTripItinerary, request.PathParameters["itinerary_id"], Itinerary{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete remove participant
func (i *Itinerary) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			dbr.Eq("i.id", request.PathParameters["itinerary_id"]),
			dbr.Eq("p.trip_id", request.PathParameters["id"]),
			dbr.Eq("p.user_id", tokenUser.UserID),
			dbr.Or(
				dbr.Eq("i.created_by", tokenUser.UserID),
				dbr.Eq("p.role", ParticipantOwnerRole),
				dbr.Eq("p.role", ParticipantAdminRole),
			),
		)
		table := db.TableTripParticipant + " p , " + db.TableTripItinerary + " i"
		total, err := db.Validate(session, []string{"count(p.id) total"}, table, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("not allowed to delete this itinerary"))
		}
	}

	err = db.Delete(session, db.TableTripItinerary, request.PathParameters["itinerary_id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
