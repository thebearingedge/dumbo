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

func InsertMany(t *testing.T, db DB, table string, records []Record) []Record {

	_, err := db.Exec(fmt.Sprintf(`truncate table %q restart identity`, table))
	require.NoError(t, err, fmt.Sprintf("truncating table %q", table))

	first := records[0]

	columns := make([]string, 0, len(first))
	for column := range first {
		columns = append(columns, fmt.Sprintf("%q", column))
	}

	params := make([]string, 0, len(records))
	values := make([]any, 0, len(records))
	p := 1

	for _, record := range records {
		tuple := make([]string, 0, len(first))
		for column := range first {
			values = append(values, record[column])
			tuple = append(tuple, fmt.Sprintf("$%v", p))
			p++
		}
		params = append(params, fmt.Sprintf("(%v)", strings.Join(tuple, ",")))
	}

	rows, err := db.Query(fmt.Sprintf(`
		insert into %v (%v)
		values %v
		returning *
	`, fmt.Sprintf("%q", table), strings.Join(columns, ", "), strings.Join(params, ", ")), values...)
	require.NoError(t, err, fmt.Sprintf("inserting row(s) into table %q", table))

	returned, err := rows.Columns()
	require.NoError(t, err, fmt.Errorf("reading columns returned from table %q", table))

	inserted := make([]Record, 0, len(records))

	for rows.Next() {
		fields := make([]any, len(returned))
		for i := range fields {
			fields[i] = &fields[i]
		}

		require.NoError(t, rows.Scan(fields...), fmt.Sprintf("scanning row returned from %q", table))

		record := make(Record, len(returned))
		for i, column := range returned {
			record[column] = fields[i]
		}

		inserted = append(inserted, record)
	}

	require.NoError(t, rows.Err(), fmt.Sprintf("iterating rows returned from table %q", table))

	return inserted
}

func InsertOne(t *testing.T, db DB, table string, record Record) Record {
	return InsertMany(t, db, table, []Record{record})[0]
}
