package resources

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/feedmytrip/api/common"
	validator "gopkg.in/go-playground/validator.v9"
)

//Auth represents the attribute to authenticate an user
type Auth struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

//Login validate user credentials with AWS Cognito
func (a *Auth) Login(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), a)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	validate := validator.New()
	err = validate.Struct(a)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	svc := cognitoidentityprovider.New(sess)
	authInput := &cognitoidentityprovider.AdminInitiateAuthInput{
		ClientId:   aws.String("2i1vka74ub2c6b3i5l6o3aaio"),
		UserPoolId: aws.String("us-east-1_0JwI28hrb"),
		AuthFlow:   aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(a.Username),
			"PASSWORD": aws.String(a.Password),
		},
	}
	authOutput, err := svc.AdminInitiateAuth(authInput)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(authOutput.AuthenticationResult, http.StatusOK)
}
