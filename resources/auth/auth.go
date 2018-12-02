package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/gbrlsnchs/jwt"
)

//TODO: set as lambda environment variables
const (
	clientID   = "2i1vka74ub2c6b3i5l6o3aaio"
	userPoolID = "us-east-1_0JwI28hrb"
)

//Auth represents the attribute to authenticate an user
type Auth struct{}

type userCredentials struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	Email        string `json:"email"`
	Group        string `json:"group"`
	LanguageCode string `json:"language_code"`
}

//UserResponse represents a response from AWS Cognito after login
type UserResponse struct {
	UserID    string                                            `json:"userId"`
	Group     string                                            `json:"group"`
	Email     string                                            `json:"email"`
	FirstName string                                            `json:"firstName"`
	LastName  string                                            `json:"lastName"`
	Tokens    *cognitoidentityprovider.AuthenticationResultType `json:"tokens"`
}

//DBUser represents the database fields dependent to the AWS Cognito
type DBUser struct {
	ID           string    `json:"id" db:"id"`
	Active       bool      `json:"active" db:"active"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Group        string    `json:"group" db:"group"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	CreatedDate  time.Time `json:"created_date" db:"created_date"`
	UpdatedBy    string    `json:"updated_by" db:"updated_by"`
	UpdatedDate  time.Time `json:"updated_date" db:"updated_date"`
}

//Register creates a new user in AWS Cognito and in the Database
func (a *Auth) Register(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admin users can access this resource"))
	}

	credentials := userCredentials{}
	err := json.Unmarshal([]byte(request.Body), &credentials)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	sess, err := getAWSSession("us-east-1")
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	svc := cognitoidentityprovider.New(sess)
	signUpParams := &cognitoidentityprovider.SignUpInput{
		Username: aws.String(credentials.Username),
		Password: aws.String(credentials.Password),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			&cognitoidentityprovider.AttributeType{
				Name:  aws.String("given_name"),
				Value: aws.String(credentials.GivenName),
			},
			&cognitoidentityprovider.AttributeType{
				Name:  aws.String("family_name"),
				Value: aws.String(credentials.FamilyName),
			},
			&cognitoidentityprovider.AttributeType{
				Name:  aws.String("email"),
				Value: aws.String(credentials.Email),
			},
			&cognitoidentityprovider.AttributeType{
				Name:  aws.String("custom:language_code"),
				Value: aws.String(credentials.LanguageCode),
			},
		},
		ClientId: aws.String(clientID),
	}

	result, err := svc.SignUp(signUpParams)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	confirmSignUpParams := &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(credentials.Username),
	}

	_, err = svc.AdminConfirmSignUp(confirmSignUpParams)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	if credentials.Group != "User" {
		addUserToGroupParams := &cognitoidentityprovider.AdminAddUserToGroupInput{
			GroupName:  aws.String(credentials.Group),
			UserPoolId: aws.String(userPoolID),
			Username:   aws.String(credentials.Username),
		}

		_, err := svc.AdminAddUserToGroup(addUserToGroupParams)
		if err != nil {
			return common.APIError(http.StatusInternalServerError, err)
		}
	}

	user := DBUser{
		ID:           *result.UserSub,
		Active:       true,
		FirstName:    credentials.GivenName,
		LastName:     credentials.FamilyName,
		Group:        credentials.Group,
		Username:     credentials.Username,
		Email:        credentials.Email,
		LanguageCode: credentials.LanguageCode,
	}
	user.CreatedBy = tokenUser.UserID
	user.CreatedDate = time.Now()
	user.UpdatedBy = tokenUser.UserID
	user.UpdatedDate = time.Now()

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

	err = db.Insert(tx, db.TableUser, user)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()
	return common.APIResponse(user, http.StatusCreated)
}

//Login validate user credentials with AWS Cognito and returns an APIGatewayProxyResponse
func (a *Auth) Login(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := LoginUser(request.Body)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}
	return common.APIResponse(user, http.StatusOK)
}

//LoginUser validate user credentials with AWS Cognito
func LoginUser(credentialsJSON string) (*UserResponse, error) {
	credentials := userCredentials{}
	err := json.Unmarshal([]byte(credentialsJSON), &credentials)
	if err != nil {
		return nil, err
	}

	if credentials.Username == "" || credentials.Password == "" {
		return nil, errors.New("empty username or password")
	}

	sess, err := getAWSSession("us-east-1")
	if err != nil {
		return nil, err
	}

	svc := cognitoidentityprovider.New(sess)
	authInput := &cognitoidentityprovider.AdminInitiateAuthInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
		AuthFlow:   aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(credentials.Username),
			"PASSWORD": aws.String(credentials.Password),
		},
	}
	authOutput, err := svc.AdminInitiateAuth(authInput)
	if err != nil {
		return nil, err
	}

	return parseUserResponse(authOutput), nil
}

//Refresh take in a valid refresh token and return new tokens
func (a *Auth) Refresh(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess, err := getAWSSession("us-east-1")
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	refreshToken := ""
	if val, ok := request.Headers["Authorization"]; ok {
		refreshToken = val
	}

	if refreshToken == "" {
		return common.APIError(http.StatusRequestHeaderFieldsTooLarge, errors.New("missing header Authorization"))
	}

	svc := cognitoidentityprovider.New(sess)
	authInput := &cognitoidentityprovider.AdminInitiateAuthInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
		AuthFlow:   aws.String("REFRESH_TOKEN"),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(refreshToken),
		},
	}
	authOutput, err := svc.AdminInitiateAuth(authInput)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	user := parseUserResponse(authOutput)
	return common.APIResponse(user, http.StatusOK)
}

type payload struct {
	*jwt.JWT
	Sub        string   `json:"sub"`
	Email      string   `json:"email"`
	GivenName  string   `json:"given_name"`
	FamilyName string   `json:"family_name"`
	Groups     []string `json:"cognito:groups"`
}

func parseUserResponse(authOutput *cognitoidentityprovider.AdminInitiateAuthOutput) *UserResponse {
	u := &UserResponse{}
	u.Tokens = authOutput.AuthenticationResult

	jwtPayload, _, _ := jwt.Parse(*u.Tokens.IdToken)
	payload := payload{}
	jwt.Unmarshal(jwtPayload, &payload)

	u.Email = payload.Email
	u.FirstName = payload.GivenName
	u.LastName = payload.FamilyName
	u.UserID = payload.Sub
	if len(payload.Groups) > 0 {
		u.Group = strings.Join(payload.Groups, ",")
	}
	return u
}

func getAWSSession(region string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, err
	}
	return sess, nil
}
