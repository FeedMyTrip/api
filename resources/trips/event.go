package trips

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	fmt "github.com/feedmytrip/api/resources/events"
	"github.com/feedmytrip/api/resources/shared"
	"github.com/gocraft/dbr"
	"github.com/google/uuid"
)

//ItineraryEvent represents an event inside a trip itinerary
type ItineraryEvent struct {
	ID                  string             `json:"id" db:"id" lock:"true"`
	TripID              string             `json:"trip_id" db:"trip_id" lock:"true"`
	ItineraryID         string             `json:"itinerary_id" db:"itinerary_id" lock:"true"`
	GlobalEventID       string             `json:"global_event_id" db:"global_event_id" lock:"true"`
	Title               shared.Translation `json:"title" table:"translation" alias:"title" on:"title.parent_id = trip_itinerary_event.id and title.field = 'title'" embedded:"true" persist:"true"`
	Description         shared.Translation `json:"description" table:"translation" alias:"description" on:"description.parent_id = trip_itinerary_event.id and description.field = 'description'" embedded:"true" persist:"true"`
	BeginOffset         float64            `json:"begin_offset" db:"begin_offset"`
	Duration            int                `json:"duration" db:"duration"`
	MainCategoryID      string             `json:"main_category_id" db:"main_category_id"`
	MainCategory        shared.Translation `json:"main_category" table:"translation" alias:"main_category" on:"main_category.parent_id = trip_itinerary_event.main_category_id and main_category.field = 'title'" embedded:"true"`
	SecondaryCategoryID string             `json:"secondary_category_id" db:"secondary_category_id"`
	SecondaryCategory   shared.Translation `json:"secondary_category" table:"translation" alias:"secondary_category" on:"secondary_category.parent_id = trip_itinerary_event.secondary_category_id and secondary_category.field = 'title'" embedded:"true"`
	CountryID           string             `json:"country_id" db:"country_id"`
	Country             shared.Translation `json:"country" table:"translation" alias:"country" on:"country.parent_id = trip_itinerary_event.country_id and country.field = 'title'" embedded:"true"`
	RegionID            string             `json:"region_id" db:"region_id"`
	Region              shared.Translation `json:"region" table:"translation" alias:"region" on:"region.parent_id = trip_itinerary_event.region_id and region.field = 'title'" embedded:"true"`
	CityID              string             `json:"city_id" db:"city_id"`
	City                shared.Translation `json:"city" table:"translation" alias:"city" on:"city.parent_id = trip_itinerary_event.city_id and city.field = 'title'" embedded:"true"`
	Address             string             `json:"address" db:"address"`
	CreatedBy           string             `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate         time.Time          `json:"created_date" db:"created_date" lock:"true"`
	UpdatedBy           string             `json:"updated_by" db:"updated_by"`
	UpdatedDate         time.Time          `json:"updated_date" db:"updated_date"`
	EvaluatedBy         string             `json:"evaluated_by" db:"evaluated_by"`
	EvaluatedDate       time.Time          `json:"evaluated_date" db:"evaluated_date"`
	EvaluatedComment    string             `json:"evaluated_comment" db:"evaluated_comment"`
	CreatedUser         shared.User        `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = trip_itinerary_event.created_by" embedded:"true"`
	UpdatedUser         shared.User        `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = trip_itinerary_event.updated_by" embedded:"true"`
	EvaluatedUser       shared.User        `json:"evaluated_user" table:"user" alias:"evaluated_user" on:"evaluated_user.id = trip_itinerary_event.evaluated_by" embedded:"true"`
}

//Get return an itinerary event
func (e *ItineraryEvent) Get(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			dbr.Eq("p.trip_id", request.PathParameters["id"]),
			dbr.Eq("p.user_id", tokenUser.UserID),
		)
		table := db.TableTripParticipant
		total, err := db.Validate(session, []string{"count(id) total"}, table, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("not allowed to view this events"))
		}
	}

	result, err := db.QueryOne(session, db.TableTripItineraryEvent, request.PathParameters["event_id"], ItineraryEvent{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//GetAll returns all itinerary events available in the database
func (e *ItineraryEvent) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			dbr.Eq("p.trip_id", request.PathParameters["id"]),
			dbr.Eq("p.user_id", tokenUser.UserID),
		)
		table := db.TableTripParticipant
		total, err := db.Validate(session, []string{"count(id) total"}, table, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("not allowed to view this events"))
		}
	}

	if request.QueryStringParameters == nil {
		request.QueryStringParameters = map[string]string{
			"trip_id":      request.PathParameters["id"],
			"itinerary_id": request.PathParameters["itinerary_id"],
		}
	} else if request.QueryStringParameters["all"] != "true" && !tokenUser.IsAdmin() {
		request.QueryStringParameters["trip_id"] = request.PathParameters["id"]
		request.QueryStringParameters["itinerary_id"] = request.PathParameters["itinerary_id"]
	}

	result, err := db.Select(session, db.TableTripItineraryEvent, request.QueryStringParameters, ItineraryEvent{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Add clone an global event to this itinerary
func (e *ItineraryEvent) Add(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			return common.APIError(http.StatusForbidden, errors.New("not allowed to delete this event"))
		}
	}

	result, err := db.QueryOne(session, db.TableEvent, request.PathParameters["global_event_id"], fmt.Event{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	jsonBytes, _ := json.Marshal(result)
	json.Unmarshal(jsonBytes, e)

	e.ID = uuid.New().String()
	e.TripID = request.PathParameters["id"]
	e.ItineraryID = request.PathParameters["itinerary_id"]
	e.GlobalEventID = request.PathParameters["global_event_id"]
	e.Title.ID = uuid.New().String()
	e.Title.Table = db.TableTripItineraryEvent
	e.Title.ParentID = e.ID
	e.Description.ID = uuid.New().String()
	e.Description.Table = db.TableTripItineraryEvent
	e.Description.ParentID = e.ID
	e.BeginOffset = -1
	e.Duration = 21600
	e.CreatedBy = tokenUser.UserID
	e.CreatedDate = time.Now()
	e.UpdatedBy = tokenUser.UserID
	e.UpdatedDate = time.Now()

	if request.Body != "" {
		jsonMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(request.Body), &jsonMap)
		if err != nil {
			return common.APIError(http.StatusBadRequest, err)
		}

		if val, ok := jsonMap["begin_offset"]; ok {
			e.BeginOffset = val.(float64)
		}
	}

	tx, err := session.Begin()
	defer tx.RollbackUnlessCommitted()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = db.Insert(tx, db.TableTripItineraryEvent, *e)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(e, http.StatusCreated)
}

//SaveNew creates a new itinerary event
func (e *ItineraryEvent) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			return common.APIError(http.StatusForbidden, errors.New("not allowed to delete this event"))
		}
	}

	err = json.Unmarshal([]byte(request.Body), e)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if e.Title.IsEmpty() {
		return common.APIError(http.StatusBadRequest, errors.New("invalid request empty title"))
	}

	e.ID = uuid.New().String()
	e.TripID = request.PathParameters["id"]
	e.ItineraryID = request.PathParameters["itinerary_id"]
	e.GlobalEventID = ""
	e.Title.ID = uuid.New().String()
	e.Title.Table = db.TableTripItineraryEvent
	e.Title.Field = "title"
	e.Title.ParentID = e.ID
	e.Description.ID = uuid.New().String()
	e.Description.Table = db.TableTripItineraryEvent
	e.Description.Field = "description"
	e.Description.ParentID = e.ID
	e.BeginOffset = -1
	e.Duration = 21600
	e.CreatedBy = tokenUser.UserID
	e.CreatedDate = time.Now()
	e.UpdatedBy = tokenUser.UserID
	e.UpdatedDate = time.Now()

	e.EvaluatedBy = ""
	e.EvaluatedComment = ""

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Insert(tx, db.TableTripItineraryEvent, *e)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(e, http.StatusCreated)
}

//Update change event attributes in the database
func (e *ItineraryEvent) Update(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			return common.APIError(http.StatusForbidden, errors.New("not allowed to delete this event"))
		}
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(request.Body), &jsonMap)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	if !tokenUser.IsAdmin() {
		delete(jsonMap, "evaluated_by")
		delete(jsonMap, "evaluated_date")
		delete(jsonMap, "evaluated_comment")
	}

	jsonMap["updated_by"] = tokenUser.UserID
	jsonMap["updated_date"] = time.Now()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Update(tx, db.TableTripItineraryEvent, request.PathParameters["event_id"], *e, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	result, err := db.QueryOne(session, db.TableTripItineraryEvent, request.PathParameters["event_id"], ItineraryEvent{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//Delete removes event from the database
func (e *ItineraryEvent) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			return common.APIError(http.StatusForbidden, errors.New("not allowed to delete this event"))
		}
	}

	err = db.Delete(session, db.TableTripItineraryEvent, request.PathParameters["event_id"])
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}

func (e *ItineraryEvent) clone(tx *dbr.Tx, tripID, itineraryID, userID string, offset float64) error {

	e.ID = uuid.New().String()
	e.TripID = tripID
	e.ItineraryID = itineraryID
	e.Title.ID = uuid.New().String()
	e.Title.ParentID = e.ID
	e.Description.ID = uuid.New().String()
	e.Description.ParentID = e.ID
	e.BeginOffset = e.BeginOffset + offset
	e.CreatedBy = userID
	e.CreatedDate = time.Now()
	e.UpdatedBy = userID
	e.UpdatedDate = time.Now()

	err := db.Insert(tx, db.TableTripItineraryEvent, *e)
	return err
}
