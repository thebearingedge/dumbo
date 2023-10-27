package schema

import (
	"database/sql"
)

type Credentials struct {
	Email    string
	Password string
}

type Registration struct {
	Email    string
	Password string
	Username string
}

type User struct {
	ID       int64
	Email    string
	Username string
	Bio      sql.NullString
	Image    sql.NullString
}
