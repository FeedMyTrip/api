package resources

import "time"

//Rating defines an object to consolidate system ratings
type Rating struct {
	RatingID    string    `json:"ratingId"`
	UserID      string    `json:"userId"`
	Score       int       `json:"score"`
	CreatedDate time.Time `json:"createdDate"`
}
