package model

import "time"

// Contact is the data structure for a person that we know.
// All fields with the exception of the Id field are optional.
type Contact struct {
	Id        int64      `json:"id"                  db:"id"`
	FirstName *string    `json:"firstname,omitempty" db:"firstname"`
	LastName  *string    `json:"lastname,omitempty"  db:"lastname"`
	Phone     *string    `json:"phone,omitempty"     db:"phone"`
	Birthday  *time.Time `json:"birthday,omitempty"  db:"birthday"`
}
