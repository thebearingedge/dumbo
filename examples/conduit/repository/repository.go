package repository

import (
	"database/sql"
)

type service interface {
	Exec(sql string, values ...any) (sql.Result, error)
	QueryRows(sql string, values ...any) (*sql.Rows, error)
}
