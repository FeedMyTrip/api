package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/translate"
	"github.com/feedmytrip/api/common"
)

//Translation represents text translated into system languages
type Translation struct {
	ID       string `json:"id" db:"id" lock:"true"`
	ParentID string `json:"parent_id" db:"parent_id" lock:"true"`
	Field    string `json:"field" db:"field" lock:"true"`
	PT       string `json:"pt" db:"pt" filter:"true"`
	ES       string `json:"es" db:"es" filter:"true"`
	EN       string `json:"en" db:"en" filter:"true"`
}

//IsEmpty returns true if the translation is empty
func (t *Translation) IsEmpty() bool {
	if t.PT == "" && t.EN == "" && t.ES == "" {
		return true
	}
	return false
}

//Translate uses AWS Translate service to translate empty languages
func (t *Translation) Translate() error {
	if t.IsEmpty() {
		return nil
	}
	var err error
	sess, _ := common.GetAWSSession("us-east-1")
	if t.EN == "" {
		if t.PT != "" {
			t.EN, err = autoTranslate(sess, "pt", "en", t.PT)
			if err != nil {
				return err
			}
		} else {
			t.EN, err = autoTranslate(sess, "es", "en", t.ES)
			if err != nil {
				return err
			}
		}
	}

	if t.PT == "" {
		t.PT, err = autoTranslate(sess, "en", "pt", t.EN)
		if err != nil {
			return err
		}
	}
	if t.ES == "" {
		t.ES, err = autoTranslate(sess, "en", "es", t.EN)
		if err != nil {
			return err
		}
	}

	return nil
}

func autoTranslate(sess *session.Session, source, target, text string) (string, error) {
	svc := translate.New(sess)
	input := &translate.TextInput{
		SourceLanguageCode: aws.String(source),
		TargetLanguageCode: aws.String(target),
		Text:               aws.String(text),
	}
	response, err := svc.Text(input)
	if err != nil {
		return "", err
	}
	return *response.TranslatedText, nil
}
