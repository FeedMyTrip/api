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

const (
	//UserTripViewScope defines the user trips that can only be viewed
	UserTripViewScope = "viewonly"
	//UserTripEditScope defines the user trips that has itineraries that can be edited
	UserTripEditScope = "editable"
	//UserTripArchiveScope defines the user trips that are not active any more
	UserTripArchiveScope = "archived"
	//UserFavoriteEventsScope defines te user favorite events
	UserFavoriteEventsScope = "events"
	//UserFavoriteTripsScope defines te user favorite trips
	UserFavoriteTripsScope = "trips"
)

//User represents a user in the system
type User struct {
	UserID        string              `json:"userId"`
	ActiveTripID  string              `json:"activeTripId"`
	Trips         map[string][]string `json:"trips"`
	Favorites     map[string][]string `json:"favorites"`
	Notifications []string            `json:"notifications"`
	CountryID     string              `json:"countryId"`
	RegionID      string              `json:"regionId"`
	CityID        string              `json:"cityId"`
	AboutMe       string              `json:"aboutme"`
}

//Notification represents a feedback from the system to a user
type Notification struct {
	NotificationID string      `json:"notificationId"`
	Text           Translation `json:"text"`
	CreatedDate    time.Time   `json:"createdDate"`
}

type userDetailsResponse struct {
	User           *User          `json:"user"`
	Trips          []Trip         `json:"trips,omitempty"`
	FavoriteTrips  []Trip         `json:"favoriteTrips,omitempty"`
	FavoriteEvents []Event        `json:"favoriteEvents,omitempty"`
	Notifications  []Notification `json:"notifications,omitempty"`
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

//ToggleFavoriteContent add or remove trip or event from users favorites
func (u *User) ToggleFavoriteContent(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	loggedUser := common.GetTokenUser(request)
	err := u.Load(loggedUser.UserID)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	contentType := request.PathParameters["contentType"]
	contentID := request.PathParameters["contentId"]

	_, index := getMapContentScopeAndIndex(u.Favorites, contentID)

	unit := 0
	if index == -1 {
		jsonMap := make(map[string]interface{})
		content := []string{}
		content = append(content, contentID)
		jsonMap[":favorites"+contentType] = content

		_, err = db.PutListItem(common.UsersTable, "userId", loggedUser.UserID, "favorites."+contentType, jsonMap)
		unit = 1
	} else {
		err = db.DeleteListItem(common.UsersTable, "userId", loggedUser.UserID, "favorites."+contentType, index)
		unit = -1
	}
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	total := 0
	table := ""
	tablePK := ""
	if contentType == UserFavoriteEventsScope {
		e := Event{}
		e.Load(contentID)
		total = e.Favorite
		table = common.EventsTable
		tablePK = "eventId"
	} else if contentType == UserFavoriteTripsScope {
		t := Trip{}
		t.Load(contentID)
		total = t.Favorite
		table = common.TripsTable
		tablePK = "tripId"
	}

	jsonMap := make(map[string]interface{})
	jsonMap["favorite"] = total + unit
	_, err = db.UpdateItem(table, tablePK, contentID, jsonMap)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(u, http.StatusOK)
}

//GetUserDetails return all users details
func (u *User) GetUserDetails(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user := common.GetTokenUser(request)
	if !user.IsAdmin() && user.UserID != request.PathParameters["id"] {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}
	u.Load(request.PathParameters["id"])
	if u.UserID == "" {
		u.UserID = request.PathParameters["id"]

		u.Trips = make(map[string][]string)
		u.Favorites = make(map[string][]string)

		//Create dummy itens to avoid sdk problem with empty map
		u.Trips[UserTripEditScope] = append(u.Trips[UserTripEditScope], "")
		u.Trips[UserTripViewScope] = append(u.Trips[UserTripViewScope], "")
		u.Trips[UserTripArchiveScope] = append(u.Trips[UserTripArchiveScope], "")
		u.Favorites[UserFavoriteEventsScope] = append(u.Favorites[UserFavoriteEventsScope], "")
		u.Favorites[UserFavoriteTripsScope] = append(u.Favorites[UserFavoriteTripsScope], "")
		u.Notifications = append(u.Notifications, "")

		err := db.PutItem(u, common.UsersTable)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}

		//Delete dummy item to avoid sdk problem with empty map
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "trips."+UserTripEditScope, 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "trips."+UserTripViewScope, 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "trips."+UserTripArchiveScope, 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "favorites."+UserFavoriteEventsScope, 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "favorites."+UserFavoriteTripsScope, 0)
		db.DeleteListItem(common.UsersTable, "userId", u.UserID, "notifications", 0)

	}

	udr := userDetailsResponse{}
	udr.User = u

	if tripScope, ok := request.QueryStringParameters["trips"]; ok {
		ids := []string{}
		if tripScope == "all" || strings.Contains(tripScope, UserTripEditScope) {
			ids = append(ids, u.Trips[UserTripEditScope]...)
		}
		if tripScope == "all" || strings.Contains(tripScope, UserTripViewScope) {
			ids = append(ids, u.Trips[UserTripViewScope]...)
		}
		if tripScope == "all" || strings.Contains(tripScope, UserTripArchiveScope) {
			ids = append(ids, u.Trips[UserTripArchiveScope]...)
		}

		list := []Trip{}
		getBatchContentDetails(ids, &list)
		udr.Trips = list
	}

	if favoriteScope, ok := request.QueryStringParameters["favorites"]; ok {
		if favoriteScope == "all" || strings.Contains(favoriteScope, UserFavoriteEventsScope) {
			list := []Event{}
			getBatchContentDetails(u.Favorites[UserFavoriteEventsScope], &list)
			udr.FavoriteEvents = list
		}
		if favoriteScope == "all" || strings.Contains(favoriteScope, UserFavoriteTripsScope) {
			list := []Trip{}
			getBatchContentDetails(u.Favorites[UserFavoriteTripsScope], &list)
			udr.FavoriteTrips = list
		}
	}

	if _, ok := request.QueryStringParameters["notifications"]; ok {
		list := []Notification{}
		getBatchContentDetails(u.Notifications, &list)
		udr.Notifications = list
	}

	return common.APIResponse(udr, http.StatusOK)
}

//AddTripToUser include a new trip to the user
func AddTripToUser(userID, tripID, scope string) error {
	RemoveTripFromUser(userID, tripID)
	jsonMap := make(map[string]interface{})
	content := []string{}
	content = append(content, tripID)
	jsonMap[":trips"+scope] = content

	_, err := db.PutListItem(common.UsersTable, "userId", userID, "trips."+scope, jsonMap)
	return err
}

//RemoveTripFromUser delete a trip from user
func RemoveTripFromUser(userID, tripID string) error {
	user := User{}
	user.Load(userID)
	scope, index := getMapContentScopeAndIndex(user.Trips, tripID)
	if index == -1 {
		return errors.New("trip id not found at user`s trips")
	}
	return db.DeleteListItem(common.UsersTable, "userId", userID, "trips."+scope, index)
}

func getBatchContentDetails(ids []string, list interface{}) {
	if len(ids) == 0 {
		return
	}
	result, _ := db.BatchGetItem(common.TripsTable, "tripId", ids)
	dynamodbattribute.UnmarshalListOfMaps(result.Responses[common.TripsTable], &list)
}

func getTripsDetails(ids []string) []Trip {
	if len(ids) == 0 {
		return []Trip{}
	}
	result, _ := db.BatchGetItem(common.TripsTable, "tripId", ids)
	trips := []Trip{}
	dynamodbattribute.UnmarshalListOfMaps(result.Responses[common.TripsTable], &trips)
	return trips
}

func getEventsDetails(ids []string) []Event {
	if len(ids) == 0 {
		return []Event{}
	}
	result, _ := db.BatchGetItem(common.EventsTable, "eventId", ids)
	events := []Event{}
	dynamodbattribute.UnmarshalListOfMaps(result.Responses[common.EventsTable], &events)
	return events
}

func getMapContentScopeAndIndex(content map[string][]string, id string) (string, int) {
	for scope, ids := range content {
		for index, contentID := range ids {
			if contentID == id {
				return scope, index
			}
			index++
		}
	}
	return "", -1
}
