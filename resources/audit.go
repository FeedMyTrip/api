package resources

import (
	"time"
)

//Audit fields to monitoring objects life cycle
type Audit struct {
	CreatedBy   string    `json:"createdBy"`
	CreatedDate time.Time `json:"createdDate"`
	UpdatedBy   string    `json:"updatedBy"`
	UpdatedDate time.Time `json:"updatedDate"`
}

//NewAudit return an poniter to an audit object filled with the userID defined
func NewAudit(userID string) *Audit {
	a := &Audit{}
	a.CreatedBy = userID
	a.CreatedDate = time.Now()
	a.UpdatedBy = userID
	a.UpdatedDate = time.Now()
	return a
}

func (a *Audit) updateAudit(userID string) {
	a.UpdatedBy = userID
	a.UpdatedDate = time.Now()
}
