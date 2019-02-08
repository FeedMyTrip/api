package shared

//Translation represents text translated into system languages
type Translation struct {
	ID       string `json:"id" db:"id" lock:"true"`
	ParentID string `json:"parent_id" db:"parent_id" lock:"true"`
	Table    string `json:"table" db:"table" lock:"true"`
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
