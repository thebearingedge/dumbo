package dumbo

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type DB interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

type Record map[string]any

type Indexer func(r Record) string

type Factory struct {
	Table     string
	NewRecord func() Record
	UniqueBy  []Indexer
}

type Index map[string]any

type Config struct {
	retries int
}

func Defaults() Config {
	return Config{
		retries: 5,
	}
}

type Dumbo struct {
	factories map[string]Factory
	runs      []map[string][]Index
	config    Config
}

func New(factories ...Factory) Dumbo {
	d := Dumbo{
		factories: make(map[string]Factory, len(factories)),
		runs:      make([]map[string][]Index, 0, 1),
		config:    Defaults(),
	}

	run := make(map[string][]Index)

	for _, factory := range factories {
		indexes := make([]Index, len(factory.UniqueBy))
		for i := range factory.UniqueBy {
			indexes[i] = make(Index)
		}
		run[factory.Table] = indexes
		d.factories[factory.Table] = factory
	}

	d.runs = append(d.runs, run)

	return d
}

func NewWithConfig(config Config, factories ...Factory) Dumbo {
	d := New(factories...)
	d.config = config
	return d
}

// Truncate the target table before inserting the record.
func (d *Dumbo) SeedOne(t *testing.T, db DB, table string, partial Record) Record {
	return d.SeedMany(t, db, table, []Record{partial})[0]
}

// Truncate the target table before inserting the records.
func (d *Dumbo) SeedMany(t *testing.T, db DB, table string, partials []Record) []Record {
	factory, hasFactory := d.factories[table]
	if !hasFactory {
		return seed(t, db, table, partials)
	}

	return seed(t, db, table, generate(d.runs, d.config.retries, factory, partials))
}

// Add a record to the target table.
func (d *Dumbo) InsertOne(t *testing.T, db DB, table string, partial Record) Record {
	return d.InsertMany(t, db, table, []Record{partial})[0]
}

// Add records to the target table.
func (d *Dumbo) InsertMany(t *testing.T, db DB, table string, partials []Record) []Record {
	factory, hasFactory := d.factories[table]
	if !hasFactory {
		return insert(t, db, table, partials)
	}

	run := d.runs[len(d.runs)-1]
	_, hasIndexes := run[table]
	if !hasIndexes {
		run[table] = make([]Index, len(factory.UniqueBy))
		for i := range factory.UniqueBy {
			run[table][i] = make(Index)
		}
	}

	return insert(t, db, table, generate(d.runs, d.config.retries, factory, partials))
}

func (d Dumbo) FetchOne(t *testing.T, db DB, query string, values ...any) Record {
	return d.FetchMany(t, db, query, values...)[0]
}

func (d Dumbo) FetchMany(t *testing.T, db DB, query string, values ...any) []Record {
	rows, err := db.Query(query, values...)
	require.NoError(t, err, fmt.Sprintf("running query %q", query))

	return fetchAll(t, rows)
}

// Remove unique indexes from sub-test when done.
func (d *Dumbo) Run(t *testing.T, r func(d *Dumbo)) {
	d.runs = append(d.runs, make(map[string][]Index))
	t.Cleanup(func() { d.runs = d.runs[:len(d.runs)-1] })
	r(d)
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

	return fetchAll(t, rows)
}

func seed(t *testing.T, db DB, table string, records []Record) []Record {
	_, err := db.Exec(fmt.Sprintf(`truncate table %q restart identity cascade`, table))
	require.NoError(t, err, fmt.Sprintf("truncating table %q", table))

	return insert(t, db, table, records)
}

func generate(runs []map[string][]Index, retries int, factory Factory, partials []Record) []Record {

	records := make([]Record, 0, len(partials))

EACH_PARTIAL:
	for _, partial := range partials {
		attempts := 0

	EACH_RECORD:
		for {
			if attempts > retries {
				err := fmt.Errorf(
					"maximum %v retries exceeded generating record for table %q",
					retries,
					factory.Table,
				)
				panic(err)
			}

			record := factory.NewRecord()
			for column, value := range partial {
				record[column] = value
			}

			for i, run := range runs {
				indexes := run[factory.Table]
				for j, uniqueBy := range factory.UniqueBy {

					key := uniqueBy(record)
					if _, exists := indexes[j][key]; exists {
						attempts++
						continue EACH_RECORD
					} else {
						if i == len(runs)-1 {
							indexes[j][key] = struct{}{}
						}
					}
				}
			}

			records = append(records, record)

			continue EACH_PARTIAL
		}
	}

	return records
}

func fetchAll(t *testing.T, rows *sql.Rows) []Record {
	columns, err := rows.Columns()
	require.NoError(t, err, "reading columns returned from query")

	fetched := make([]Record, 0)

	for rows.Next() {
		fields := make([]any, len(columns))
		for i := range fields {
			fields[i] = &fields[i]
		}

		require.NoError(t, rows.Scan(fields...), "scanning row returned from query")

		record := make(Record, len(columns))
		for i, column := range columns {
			record[column] = fields[i]
		}

		fetched = append(fetched, record)
	}

	require.NoError(t, rows.Err(), "iterating rows returned from query")

	return fetched
}
