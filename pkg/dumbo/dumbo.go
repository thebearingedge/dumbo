package dumbo

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

type DB interface {
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	Exec(string, ...any) (sql.Result, error)
}

type Record map[string]any

func InsertOne(t *testing.T, db DB, table string, record Record) Record {

	_, err := db.Exec(fmt.Sprintf(`truncate table %q restart identity`, table))
	require.NoError(t, err)

	columns := make([]string, 0, len(record))
	params := make([]string, 0, len(record))
	values := make([]any, 0, len(record))
	for column := range record {
		columns = append(columns, fmt.Sprintf("%q", column))
		params = append(params, fmt.Sprintf("$%v", len(columns)))
		values = append(values, record[column])
	}

	rows, err := db.Query(fmt.Sprintf(`
		insert into %v (%v)
		values (%v)
		returning *
	`, fmt.Sprintf("%q", table), strings.Join(columns, ", "), strings.Join(params, ", ")), values...)
	require.NoError(t, err)

	returnedColumns, err := rows.Columns()
	require.NoError(t, err)

	rows.Next()
	require.NoError(t, rows.Err())

	returnedValues := make([]any, len(returnedColumns))
	for i := range returnedValues {
		returnedValues[i] = &returnedValues[i]
	}

	err = rows.Scan(returnedValues...)
	require.NoError(t, err)

	inserted := make(Record, len(returnedColumns))
	for i, column := range returnedColumns {
		inserted[column] = returnedValues[i]
	}

	return inserted
}
