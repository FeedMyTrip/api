package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/feedmytrip/api/common"
	"github.com/gbrlsnchs/jwt"
	validator "gopkg.in/go-playground/validator.v9"
)

const (
	clientID   = "2i1vka74ub2c6b3i5l6o3aaio"
	userPoolID = "us-east-1_0JwI28hrb"
)

//Auth represents the attribute to authenticate an user
type Auth struct{}

type userCredentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

//AuthUserResponse represents a response from AWS Cognito after login
type AuthUserResponse struct {
	UserID    string                                            `json:"userId"`
	Group     string                                            `json:"group"`
	Email     string                                            `json:"email"`
	FirstName string                                            `json:"firstName"`
	LastName  string                                            `json:"lastName"`
	Tokens    *cognitoidentityprovider.AuthenticationResultType `json:"tokens"`
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
func LoginUser(credentialsJSON string) (*AuthUserResponse, error) {
	credentials := userCredentials{}
	err := json.Unmarshal([]byte(credentialsJSON), &credentials)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(credentials)
	if err != nil {
		return nil, err
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

func parseUserResponse(authOutput *cognitoidentityprovider.AdminInitiateAuthOutput) *AuthUserResponse {
	u := &AuthUserResponse{}
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
