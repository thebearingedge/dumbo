package service

import (
	"database/sql"
)

type DB interface {
	Exec(sql string, values ...any) (sql.Result, error)
	Query(sql string, values ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}
