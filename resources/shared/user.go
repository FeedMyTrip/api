package shared

//User represents a simplifyed user to deal with creted_by, updated_by, owner and other user fields
type User struct {
	ID        string `json:"id" db:"id" lock:"true"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
	ImagePath string `json:"image_path" db:"image_path"`
}
