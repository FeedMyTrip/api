package trips

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources/shared"
	"github.com/gocraft/dbr"
	"github.com/google/uuid"
)

//Invite represents an invite to a Trip
type Invite struct {
	ID          string      `json:"id" db:"id" lock:"true"`
	TripID      string      `json:"trip_id" db:"trip_id" lock:"true"`
	Email       string      `json:"email" db:"email" lock:"true"`
	CreatedBy   string      `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time   `json:"created_date" db:"created_date" lock:"true"`
	CreatedUser shared.User `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = trip_invite.created_by" embedded:"true"`
}

//GetAll returns all itineraries from the trip
func (i *Invite) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	result, err := db.Select(session, db.TableTripInvite, request.QueryStringParameters, Invite{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new invite
func (i *Invite) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			return common.APIError(http.StatusForbidden, errors.New("only trip owner or admins can create invites"))
		}
	}

	err = json.Unmarshal([]byte(request.Body), i)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	if i.Email == "" || !re.MatchString(i.Email) {
		return common.APIError(http.StatusBadRequest, errors.New("invalid email"))
	}

	i.ID = uuid.New().String()
	i.TripID = request.PathParameters["id"]
	i.CreatedBy = tokenUser.UserID
	i.CreatedDate = time.Now()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Insert(tx, db.TableTripInvite, *i)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	//TODO: send invite to email

	return common.APIResponse(i, http.StatusCreated)
}

//Delete remove participant
func (i *Invite) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			dbr.Eq("i.id", request.PathParameters["invite_id"]),
			dbr.Eq("p.trip_id", request.PathParameters["id"]),
			dbr.Eq("p.user_id", tokenUser.UserID),
			dbr.Or(
				dbr.Eq("i.created_by", tokenUser.UserID),
				dbr.Eq("p.role", ParticipantOwnerRole),
				dbr.Eq("p.role", ParticipantAdminRole),
			),
		)
		table := db.TableTripParticipant + " p , " + db.TableTripInvite + " i"
		total, err := db.Validate(session, []string{"count(p.id) total"}, table, filter)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
		if total <= 0 {
			return common.APIError(http.StatusForbidden, errors.New("not allowed to delete this itinerary"))
		}
	}

	err = db.Delete(session, db.TableTripInvite, request.PathParameters["invite_id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
