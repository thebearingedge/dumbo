package repository

import (
	"database/sql"
)

type repository interface {
	QueryRows(sql string, values ...any) (*sql.Rows, error)
}
