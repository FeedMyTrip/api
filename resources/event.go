package resources

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

//Event represents an event on the system
type Event struct {
	EventID string `json:"eventId"`
	Audit   *Audit `json:"audit"`
}

//GetAll returns all events on the system
func (e *Event) GetAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotImplemented,
	}, nil
}

//SaveNew creates a new event on the system
func (e *Event) SaveNew(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotImplemented,
	}, nil
}
