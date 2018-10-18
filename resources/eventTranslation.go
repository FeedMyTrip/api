package resources

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/feedmytrip/api/common"
	"github.com/feedmytrip/api/db"
	validator "gopkg.in/go-playground/validator.v9"
)

//EventTranslation defines an translation to a defined language code
type EventTranslation struct {
	Code        string `json:"code" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

//Save update an existing translation on the database
func (et *EventTranslation) Save(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := json.Unmarshal([]byte(request.Body), et)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	validate := validator.New()
	err = validate.Struct(et)
	if err != nil {
		return common.APIError(http.StatusBadRequest, err)
	}

	e := Event{}
	e.Load(request)

	index, err := getEventTranslationIndex(e.Translations, et.Code)
	if err != nil {
		return common.APIError(http.StatusNotFound, err)
	}

	_, err = db.UpdateListItem(common.EventsTable, "eventId", e.EventID, "translations", index, et)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	e.Translations[index] = *et

	err = UpdateEventAudit(request)
	if err != nil {
		return common.APIError(http.StatusInternalServerError, err)
	}

	return common.APIResponse(e, http.StatusOK)
}

//DefaultTranslations used to populate event with default translations
func DefaultTranslations(et EventTranslation) []EventTranslation {
	translations := []EventTranslation{}
	languages := []string{"pt", "en", "es"}
	for _, l := range languages {
		if et.Code != l {
			//TODO Translate using AWS translation
			newET := EventTranslation{
				Code:        l,
				Title:       et.Title,
				Description: et.Description,
			}
			translations = append(translations, newET)
		} else {
			translations = append(translations, et)
		}
	}
	return translations
}

func getEventTranslationIndex(translations []EventTranslation, code string) (int, error) {
	index := 0
	found := false

	for _, t := range translations {
		if t.Code == code {
			found = true
			break
		}
		index++
	}

	if !found {
		return -1, errors.New("languange code not found")
	}

	return index, nil
}
