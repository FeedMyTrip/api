package resources

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
