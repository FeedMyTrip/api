package resources

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
)

//User represents a user in the system
type User struct {
	UserID         string         `json:"userId"`
	Trips          []string       `json:"trips"`
	FavoriteEvents []string       `json:"favoriteEvents"`
	FavoriteTrips  []string       `json:"favoriteTrips"`
	Notifications  []Notification `json:"notifications"`
	CountryID      string         `json:"countryId"`
	RegionID       string         `json:"regionId"`
	CityID         string         `json:"cityId"`
	AboutMe        string         `json:"aboutme"`
}

//Notification represents a feedback from the system to a user
type Notification struct {
	NotificationID string      `json:"notificationId"`
	Text           Translation `json:"text"`
	CreatedDate    time.Time   `json:"createdDate"`
}

type userDetailsResponse struct {
	User           *User   `json:"user"`
	Trips          []Trip  `json:"trips,omitempty"`
	FavoriteTrips  []Trip  `json:"favoriteTrips,omitempty"`
	FavoriteEvents []Event `json:"favoriteEvents,omitempty"`
}

//Load get user information from the database
func (u *User) Load(id string) error {
	result, err := db.GetItem(common.UsersTable, "userId", id)
	if err != nil {
		return err
	}
	err = dynamodbattribute.UnmarshalMap(result.Item, u)
	if err != nil {
		return err
	}
	return nil
}

//GetUserDetails return all users details
func (u *User) GetUserDetails(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user := common.GetTokenUser(request)
	if !user.IsAdmin() && user.UserID != request.PathParameters["id"] {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	err := u.Load(request.PathParameters["id"])
	if u.UserID == "" {
		u.UserID = request.PathParameters["id"]
		u.Trips = append(u.Trips, "")
		u.FavoriteTrips = append(u.FavoriteTrips, "")
		u.FavoriteEvents = append(u.FavoriteEvents, "")
		u.Notifications = append(u.Notifications, Notification{})

		err = db.PutItem(u, common.UsersTable)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}

		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "trips", 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "favoriteTrips", 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "favoriteEvents", 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "notifications", 0)

		return common.APIResponse(u, http.StatusOK)
	}

	udr := userDetailsResponse{}
	udr.User = u

	if include, ok := request.QueryStringParameters["include"]; ok {
		if strings.Contains(include, "trips") || strings.Contains(include, "all") {
			udr.Trips = getTripsDetails(u.Trips, "title,desciption,itineraries,itineraryId,tripId")
		}
		if strings.Contains(include, "fav-trips") || strings.Contains(include, "all") {
			udr.FavoriteTrips = getTripsDetails(u.Trips, "title,desciption,itineraries,itineraryId,tripId")
		}
		if strings.Contains(include, "fav-events") || strings.Contains(include, "all") {
			udr.FavoriteEvents = getEventsDetails(u.Trips)
		}
	}

	return common.APIResponse(udr, http.StatusOK)
}

//AddTripToUser include a new trip to the user
func AddTripToUser(userID, tripID string) error {
	jsonMap := make(map[string]interface{})
	content := []string{}
	content = append(content, tripID)
	jsonMap[":trips"] = content

	_, err := db.PutListItem(common.UsersTable, "userId", userID, "trips", jsonMap)
	return err
}

//RemoveTripFromUser delete a trip from user
func RemoveTripFromUser(userID, tripID string) error {
	user := User{}
	user.Load(userID)
	index := common.GetContentIndex(user.Trips, tripID)
	if index == -1 {
		return errors.New("trip id not found at user`s trips")
	}
	return db.DeleteListItem(common.UsersTable, "userId", userID, "trips", index)
}

func getTripsDetails(ids []string, projection string) []Trip {
	if len(ids) == 0 {
		return nil
	}
	result, _ := db.BatchGetItem(common.TripsTable, "tripId", projection, ids)
	trips := []Trip{}
	dynamodbattribute.UnmarshalListOfMaps(result.Responses[common.TripsTable], &trips)
	return trips
}

func getEventsDetails(ids []string) []Event {
	return nil
}
