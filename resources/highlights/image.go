package highlights

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	"github.com/feedmytrip/api/resources/shared"
	"github.com/google/uuid"
)

//HighlightImage represents an image in a highlight
type HighlightImage struct {
	ID          string      `json:"id" db:"id" lock:"true"`
	HighlightID string      `json:"highlight_id" db:"highlight_id" lock:"true"`
	Path        string      `json:"path" db:"path" filter:"true" lock:"true"`
	FileName    string      `json:"file_name" db:"file_name" filter:"true" lock:"true"`
	CreatedBy   string      `json:"created_by" db:"created_by" lock:"true"`
	CreatedDate time.Time   `json:"created_date" db:"created_date" lock:"true"`
	CreatedUser shared.User `json:"created_user" table:"user" alias:"created_user" on:"created_user.id = highlight_image.created_by" embedded:"true"`
}

//GetAll returns all highlight images
func (h *HighlightImage) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	result, err := db.Select(session, db.TableHighlightImage, request.QueryStringParameters, HighlightImage{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(result, http.StatusOK)
}

//SaveNew creates a new highlight image
func (h *HighlightImage) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admins can create highlight images"))
	}

	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	err = json.Unmarshal([]byte(request.Body), h)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	h.ID = uuid.New().String()
	h.HighlightID = request.PathParameters["id"]
	h.CreatedBy = tokenUser.UserID
	h.CreatedDate = time.Now()

	tx, err := session.Begin()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}
	defer tx.RollbackUnlessCommitted()

	err = db.Insert(tx, db.TableHighlightImage, *h)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	tx.Commit()

	return common.APIResponse(h, http.StatusCreated)
}

//Delete remove highlight image
func (h *HighlightImage) Delete(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tokenUser := common.GetTokenUser(request)
	if !tokenUser.IsAdmin() {
		return common.APIError(http.StatusForbidden, errors.New("only admins can delete highlight image"))
	}

	conn, err := db.Connect()
	defer conn.Close()
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	session := conn.NewSession(nil)
	defer session.Close()

	result, err := db.QueryOne(session, db.TableHighlightImage, request.PathParameters["image_id"], Highlight{})
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	jsonBytes, _ := json.Marshal(result)
	json.Unmarshal(jsonBytes, h)

	sess, err := common.GetAWSSession("us-east-1")
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	svc := s3.New(sess)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String("fmt-files"),
		Key:    aws.String(h.Path),
	}

	result, err = svc.DeleteObject(input)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	err = db.Delete(session, db.TableHighlightImage, request.PathParameters["image_id"])
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(nil, http.StatusOK)
}
