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
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
}

type Record map[string]any

func insertMany(t *testing.T, db DB, table string, records []Record) []Record {
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
		params = append(params, fmt.Sprintf("(%v)", strings.Join(tuple, ", ")))
	}

	rows, err := db.Query(fmt.Sprintf(`
		insert into %v (%v)
		values %v
		returning *
	`, fmt.Sprintf("%q", table), strings.Join(columns, ", "), strings.Join(params, ", ")), values...)
	require.NoError(t, err, fmt.Sprintf("inserting row(s) into table %q", table))

	returned, err := rows.Columns()
	require.NoError(t, err, fmt.Sprintf("reading columns returned from table %q", table))

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

func seedMany(t *testing.T, db DB, table string, records []Record) []Record {
	_, err := db.Exec(fmt.Sprintf(`truncate table %q restart identity cascade`, table))
	require.NoError(t, err, fmt.Sprintf("truncating table %q", table))

	return insertMany(t, db, table, records)
}

type Factory struct {
	Table     string
	NewRecord func() Record
}

type Seeder struct {
	factories map[string]func() Record
}

func NewSeeder(factories ...Factory) Seeder {
	factoryMap := make(map[string]func() Record, len(factories))
	for _, factory := range factories {
		factoryMap[factory.Table] = factory.NewRecord
	}

	return Seeder{
		factories: factoryMap,
	}
}

// Truncate the target table before inserting the record.
func (s *Seeder) SeedOne(t *testing.T, db DB, table string, partial Record) Record {
	return s.SeedMany(t, db, table, []Record{partial})[0]
}

// Truncate the target table before inserting the records.
func (s *Seeder) SeedMany(t *testing.T, db DB, table string, partials []Record) []Record {
	factory, ok := s.factories[table]
	if !ok {
		t.Fatal(fmt.Errorf("unknown table %q", table))
	}

	records := make([]Record, 0, len(partials))
	for _, partial := range partials {
		record := factory()
		for column, value := range partial {
			record[column] = value
		}
		records = append(records, record)
	}

	return seedMany(t, db, table, records)
}

// Add a record to the target table.
func (s *Seeder) InsertOne(t *testing.T, db DB, table string, partial Record) Record {
	return s.InsertMany(t, db, table, []Record{partial})[0]
}

// Add records to the target table.
func (s *Seeder) InsertMany(t *testing.T, db DB, table string, partials []Record) []Record {
	return insertMany(t, db, table, partials)
}
