package trips

import (
	"encoding/json"
	"errors"
	"math"
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
	CreatedUser shared.User        `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = trip_itinerary.created_by" embedded:"true"`
	UpdatedUser shared.User        `json:"updated_user" table:"user" alias:"updated_user" on:"updated_user.id = trip_itinerary.updated_by" embedded:"true"`
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

	if request.QueryStringParameters == nil {
		request.QueryStringParameters = map[string]string{
			"trip_id": request.PathParameters["id"],
		}
	} else {
		request.QueryStringParameters["trip_id"] = request.PathParameters["id"]
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

//Append include an existing itinerary to this one
func (i *Itinerary) Append(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	result, err := db.QueryOne(session, db.TableTripItinerary, request.PathParameters["itinerary_id"], Itinerary{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	itineraryBytes, _ := json.Marshal(result)
	json.Unmarshal(itineraryBytes, i)

	result, err = db.QueryOne(session, db.TableTripItinerary, request.PathParameters["append_itinerary_id"], Itinerary{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	appendItinerary := &Itinerary{}
	itineraryBytes, _ = json.Marshal(result)
	json.Unmarshal(itineraryBytes, appendItinerary)

	itineraryOffset := float64(math.Floor((i.EndDate.Sub(i.StartDate).Hours())/24+1) * 86400)

	appendItinerarytotalDays := int(math.Floor((appendItinerary.EndDate.Sub(appendItinerary.StartDate).Hours())/24 + 1))

	i.EndDate = i.EndDate.AddDate(0, 0, appendItinerarytotalDays)

	//get appended itinerary events
	filter := map[string]string{
		"trip_id":      appendItinerary.TripID,
		"itinerary_id": appendItinerary.ID,
		"results":      "1000",
	}
	result, err = db.Select(session, db.TableTripItineraryEvent, filter, ItineraryEvent{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	data := &resultItinerary{}
	dataBytes, _ := json.Marshal(result)
	json.Unmarshal(dataBytes, data)

	if len(data.Events) <= 0 {
		return common.APIResponse(nil, http.StatusOK)
	}

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	for _, e := range data.Events {
		err := e.clone(tx, request.PathParameters["id"], request.PathParameters["itinerary_id"], tokenUser.UserID, itineraryOffset)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
	}

	jsonMapUpdate := make(map[string]interface{})
	jsonMapUpdate["id"] = request.PathParameters["itinerary_id"]
	jsonMapUpdate["trip_id"] = request.PathParameters["id"]
	jsonMapUpdate["end_date"] = i.EndDate
	jsonMapUpdate["updated_date"] = time.Now()

	err = db.Update(tx, db.TableTripItinerary, request.PathParameters["itinerary_id"], *i, jsonMapUpdate)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	return common.APIResponse(nil, http.StatusOK)
}

type resultItinerary struct {
	Events []ItineraryEvent `json:"data"`
}

//SwapDay change itinerary events offset to rearrange days
func (i *Itinerary) SwapDay(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	fromDay := jsonMap["from"].(float64)
	toDay := jsonMap["to"].(float64)

	if fromDay <= 0 || toDay <= 0 || fromDay == toDay {
		return common.APIError(http.StatusBadRequest, errors.New("invalid from and/or to body attributes"))
	}

	sourceOffset := (fromDay - 1) * 86400
	targetOffset := (toDay - 1) * 86400

	//get appended itinerary events
	filter := map[string]string{
		"trip_id":      request.PathParameters["id"],
		"itinerary_id": request.PathParameters["itinerary_id"],
		"results":      "1000",
		"sort":         "begin_offset",
	}
	result, err := db.Select(session, db.TableTripItineraryEvent, filter, ItineraryEvent{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	data := &resultItinerary{}
	dataBytes, _ := json.Marshal(result)
	json.Unmarshal(dataBytes, data)

	if len(data.Events) <= 0 {
		return common.APIResponse(nil, http.StatusOK)
	}

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	for _, e := range data.Events {
		update := false
		if e.BeginOffset >= targetOffset && e.BeginOffset < sourceOffset {
			//swift one day forward
			e.BeginOffset = e.BeginOffset + 86400
			update = true
		} else if e.BeginOffset >= sourceOffset && e.BeginOffset < sourceOffset+86400 {
			//subtract targetOffset
			e.BeginOffset = e.BeginOffset - (sourceOffset - targetOffset)
			update = true
		} else if e.BeginOffset < targetOffset+86400 && e.BeginOffset >= sourceOffset+86400 {
			//swift one day backward
			e.BeginOffset = e.BeginOffset - 86400
			update = true
		}

		if update {
			jsonMap := make(map[string]interface{})
			jsonMap["trip_id"] = e.TripID
			jsonMap["itinerary_id"] = e.ItineraryID
			jsonMap["id"] = e.ID
			jsonMap["begin_offset"] = e.BeginOffset
			jsonMap["updated_by"] = tokenUser.UserID
			jsonMap["updated_date"] = time.Now()
			err = db.Update(tx, db.TableTripItineraryEvent, e.ID, e, jsonMap)
			if err != nil {
				return common.APIError(http.StatusInternalServerError, err)
			}
		}
	}

	tx.Commit()

	return common.APIResponse(nil, http.StatusOK)
}

//Delete remove itinerary
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
