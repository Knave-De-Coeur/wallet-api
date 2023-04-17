package pkg

import (
	"database/sql"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName          string       `json:"first_name"`
	LastName           string       `json:"last_name"`
	Email              string       `json:"email"`
	Age                int8         `json:"age"`
	Username           string       `json:"username"`
	Password           string       `json:"password"`
	LastLoginTimeStamp sql.NullTime `json:"-"`
}
