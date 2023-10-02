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

type Indexer func(r Record) string

type Factory struct {
	Table     string
	NewRecord func() Record
	UniqueBy  []Indexer
}

type Seeder struct {
	factories map[string]Factory
}

func NewSeeder(factories ...Factory) Seeder {
	seeder := Seeder{
		factories: make(map[string]Factory, len(factories)),
	}

	for _, factory := range factories {
		seeder.factories[factory.Table] = factory
	}

	return seeder
}

// Truncate the target table before inserting the record.
func (s *Seeder) SeedOne(t *testing.T, db DB, table string, partial Record) Record {
	return s.SeedMany(t, db, table, []Record{partial})[0]
}

// Truncate the target table before inserting the records.
func (s *Seeder) SeedMany(t *testing.T, db DB, table string, partials []Record) []Record {
	factory, hasFactory := s.factories[table]
	if !hasFactory {
		return seed(t, db, table, partials)
	}

	return seed(t, db, table, generate(table, factory, partials))
}

// Add a record to the target table.
func (s *Seeder) InsertOne(t *testing.T, db DB, table string, partial Record) Record {
	return s.InsertMany(t, db, table, []Record{partial})[0]
}

// Add records to the target table.
func (s *Seeder) InsertMany(t *testing.T, db DB, table string, partials []Record) []Record {
	factory, hasFactory := s.factories[table]
	if !hasFactory {
		return insert(t, db, table, partials)
	}

	return insert(t, db, table, generate(table, factory, partials))
}

func insert(t *testing.T, db DB, table string, records []Record) []Record {
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

func seed(t *testing.T, db DB, table string, records []Record) []Record {
	_, err := db.Exec(fmt.Sprintf(`truncate table %q restart identity cascade`, table))
	require.NoError(t, err, fmt.Sprintf("truncating table %q", table))

	return insert(t, db, table, records)
}

func generate(table string, factory Factory, partials []Record) []Record {

	indexes := make([]map[string]interface{}, len(factory.UniqueBy))
	for i := range indexes {
		indexes[i] = make(map[string]interface{})
	}

	records := make([]Record, 0, len(partials))

EACH_PARTIAL:
	for _, partial := range partials {
		retries := 5

	EACH_RECORD:
		for {
			if retries < 1 {
				panic(fmt.Errorf("maximum %v retries exceeded generating record for table %q", 5, table))
			}

			record := factory.NewRecord()
			for column, value := range partial {
				record[column] = value
			}

			for i, uniqueBy := range factory.UniqueBy {
				key := uniqueBy(record)
				if _, exists := indexes[i][key]; exists {
					retries--
					continue EACH_RECORD
				} else {
					indexes[i][key] = struct{}{}
				}
			}

			records = append(records, record)

			continue EACH_PARTIAL
		}
	}

	return records
}
