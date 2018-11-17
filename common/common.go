package common

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gbrlsnchs/jwt"
)

var (
	//TripsTable defines the database table to store the Trips and nested objects
	TripsTable = "Trips"
	//EventsTable defines the database table to store the Events and nested objects
	EventsTable = "Events"
	//UsersTable defines the database table to store the users nested objects
	UsersTable = "Users"
	//CategoriesTable defines the database table to store the Categories and nested objects
	CategoriesTable = "Categories"
	//CountriesTable defines the database table to store the Countries and Cities
	CountriesTable = "Countries"
	//GeonamesTable defines the database table to store countries and cities
	GeonamesTable = "Geonames"
	//HighlightsTable defines the database table to store cards grouping events and trips
	HighlightsTable = "Highlights"
)

//ParseRequestFilters process the request to parse the querystrings to dynamodb filters
func ParseRequestFilters(request events.APIGatewayProxyRequest) (string, map[string]*dynamodb.AttributeValue) {
	if len(request.QueryStringParameters) == 0 {
		return "", nil
	}
	filterExpression := ""
	data := make(map[string]interface{})
	for k, v := range request.QueryStringParameters {
		if k == "state" && v == "active" {
			filterExpression += "active = :active"
			data[":active"] = true
		} else if k == "limit" {
			data["limit"] = v
		} else {
			filterExpression += k + " = :" + k
			i, err := strconv.Atoi(v)
			if err == nil {
				data[":"+k] = i
			} else {
				data[":"+k] = v
			}
		}
		filterExpression += ", "
	}

	filterExpression = filterExpression[:len(filterExpression)-2]
	filterValues, _ := dynamodbattribute.MarshalMap(data)
	return filterExpression, filterValues
}

//TokenUser represents user information form the cognito token in Authorization header
type TokenUser struct {
	*jwt.JWT
	UserID       string   `json:"sub"`
	Groups       []string `json:"cognito:groups"`
	LanguageCode string   `json:"custom:language_code"`
}

//IsAdmin verify if the user is in the Admin group
func (t *TokenUser) IsAdmin() bool {
	return strings.Contains(strings.Join(t.Groups, ","), "Admin")
}

//GetTokenUser return the userID and Groups from the request access token
func GetTokenUser(request events.APIGatewayProxyRequest) *TokenUser {
	tokenUser := &TokenUser{}
	jwtPayload, _, _ := jwt.Parse(request.Headers["Authorization"])
	jwt.Unmarshal(jwtPayload, tokenUser)
	return tokenUser
}

//IsTokenUserAdmin check if token user is from the Admin group
func IsTokenUserAdmin(request events.APIGatewayProxyRequest) bool {
	user := GetTokenUser(request)
	if len(user.Groups) <= 0 {
		return false
	}
	return strings.Contains(strings.Join(user.Groups, ","), "Admin")
}

type apiErr struct {
	APIError string `json:"error"`
}

//APIError generates an api error message response with the defines error and status code
func APIError(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	e := apiErr{}
	e.APIError = err.Error()
	jsonBytes, err := json.Marshal(e)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonBytes),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}, nil
}

//APIResponse gernerates an APIGatewayProxyResponse based on the interface object
func APIResponse(object interface{}, statuscode int) (events.APIGatewayProxyResponse, error) {
	jsonObjectStr := "{}"
	if object != nil {
		jsonObject, err := json.Marshal(object)
		if err != nil {
			return APIError(http.StatusUnprocessableEntity, err)
		}
		jsonObjectStr = string(jsonObject)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statuscode,
		Body:       jsonObjectStr,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}, nil
}

//GetContentIndex return a index from a string array comparint to the string value
func GetContentIndex(content []string, id string) int {
	index := 0
	for _, c := range content {
		if c == id {
			return index
		}
		index++
	}
	return -1
}

//GetAWSSession returns a session to use with AWS Services
func GetAWSSession(region string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, err
	}
	return sess, nil
}
