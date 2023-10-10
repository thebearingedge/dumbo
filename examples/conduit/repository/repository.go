package repository

import (
	"database/sql"
)

type service interface {
	QueryRows(sql string, values ...any) (*sql.Rows, error)
}
