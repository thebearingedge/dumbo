package model

import "database/sql"

type Credentials struct {
	Email    string
	Password string
}

type Registration struct {
	Email    string
	Password string
	Username string
}

type Authentication struct {
	ID       uint64
	Email    string
	Token    string
	Username string
	Bio      sql.NullString
	Image    sql.NullString
}

type Profile struct {
	Username  string
	Bio       string
	Image     string
	Following bool
}
