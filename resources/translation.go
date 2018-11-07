package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/translate"
)

//Translation represents a string translated into multiple languages
type Translation struct {
	PT string `json:"pt"`
	ES string `json:"es"`
	EN string `json:"en"`
}

//IsEmpty returns true if the translation is empty
func (t *Translation) IsEmpty() bool {
	if t.PT == "" && t.EN == "" && t.ES == "" {
		return true
	}
	return false
}

//Translate uses AWS Translate service to translate empty languages
func (t *Translation) Translate() {
	if t.IsEmpty() {
		return
	}
	sess, _ := getAWSSession("us-east-1")
	if t.EN == "" {
		if t.PT != "" {
			t.EN, _ = autoTranslate(sess, "pt", "en", t.PT)
		} else {
			t.EN, _ = autoTranslate(sess, "es", "en", t.ES)
		}
	}

	if t.PT == "" {
		t.PT, _ = autoTranslate(sess, "en", "pt", t.EN)
	}
	if t.ES == "" {
		t.ES, _ = autoTranslate(sess, "en", "es", t.EN)
	}
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
