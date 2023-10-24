package dumbo

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/require"
)

func init() {
	strcase.ConfigureAcronym("id", "ID")
}

type DB interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

type Record map[string]any

type Generator func() Record

type Index map[string]any

type Generation []Index

type Indexer func(Record) string

type Dumbo struct {
	factories   map[string]Factory
	generations map[string][]Generation
}

type Factory struct {
	Table     string
	Generator Generator
	UniqueBy  []Indexer
}

func New(factories ...Factory) Dumbo {
	f := make(map[string]Factory, len(factories))
	g := make(map[string][]Generation)

	for _, factory := range factories {
		f[factory.Table] = factory
		g[factory.Table] = make([]Generation, 0)
	}

	return Dumbo{
		factories:   f,
		generations: g,
	}
}

func (d Dumbo) Seed(t *testing.T, db DB, table string, partials any) {
	t.Helper()

	if table == "" || strings.TrimSpace(table) == "" {
		t.Fatalf("table name must not be empty: got %q", table)
	}

	if _, ok := d.factories[table]; !ok {
		t.Fatalf("table has no associated factory: %v", table)
	}

	_, err := db.Exec(fmt.Sprintf(`truncate table %v restart identity cascade`, identifier(table)))
	require.NoError(t, err, fmt.Sprintf("truncating table %v", table))

	d.Insert(t, db, table, partials)
}

func (d *Dumbo) Insert(t *testing.T, db DB, table string, partials any) {
	t.Helper()

	factory := d.factories[table]
	generation := make([]Index, 0, len(factory.UniqueBy))
	for range factory.UniqueBy {
		generation = append(generation, make(Index))
	}

	d.generations[table] = append(d.generations[table], generation)
	t.Cleanup(func() { d.generations[table] = d.generations[table][:len(d.generations)-1] })

	// todo: validate that `partials` is a pointer or slice of pointers
	// alternatively, there could be some magic for a nicer API, but risky

	ps := many(partials)

	records, indexed, err := generate(5 /* todo: parameterize retries */, factory, d.generations[table], ps)
	t.Cleanup(func() {
		for i, key := range indexed {
			delete(generation[i], key)
		}
	})
	if err != nil {
		panic(fmt.Errorf("generating records: %w", err))
	}

	inserted := insert(t, db, table, records)

	for i, r := range inserted {
		p := reflect.Indirect(reflect.ValueOf(ps[i]))
		for recordKey, recordVal := range r {
			structKey := p.FieldByName(strcase.ToCamel(recordKey))
			if !structKey.IsValid() {
				// the target struct does not include this record field k: skipping.
				// this may be because the table includes a column that is not exposed to the app
				// todo: validate record fields sooner
				continue
			}
			var structVal any
			switch structKey.Interface().(type) {
			case sql.NullString:
				var s string
				if recordVal != nil {
					s = recordVal.(string)
				}
				structVal = sql.NullString{String: s, Valid: recordVal != nil}
			case sql.NullInt64:
				var i int64
				if recordVal != nil {
					i = recordVal.(int64)
				}
				structVal = sql.NullInt64{Int64: i, Valid: recordVal != nil}
			case sql.NullInt32:
				var i int32
				if recordVal != nil {
					i = recordVal.(int32)
				}
				structVal = sql.NullInt32{Int32: i, Valid: recordVal != nil}
			case sql.NullInt16:
				var i int16
				if recordVal != nil {
					i = recordVal.(int16)
				}
				structVal = sql.NullInt16{Int16: i, Valid: recordVal != nil}
			case sql.NullByte:
				var b byte
				if recordVal != nil {
					b = recordVal.(byte)
				}
				structVal = sql.NullByte{Byte: b, Valid: recordVal != nil}
			case sql.NullFloat64:
				var f float64
				if recordVal != nil {
					f = recordVal.(float64)
				}
				structVal = sql.NullFloat64{Float64: f, Valid: recordVal != nil}
			case sql.NullBool:
				var b bool
				if recordVal != nil {
					b = recordVal.(bool)
				}
				structVal = sql.NullBool{Bool: b, Valid: recordVal != nil}
			case sql.NullTime:
				var t time.Time
				if recordVal != nil {
					t = recordVal.(time.Time)
				}
				structVal = sql.NullTime{Time: t, Valid: recordVal != nil}
			default:
				structVal = recordVal
			}
			structKey.Set(reflect.ValueOf(structVal))
		}
	}
}

func many(partials any) []any {
	ps := make([]any, 0)
	p := reflect.ValueOf(partials)
	if reflect.Indirect(p).Kind() == reflect.Slice {
		s := reflect.Indirect(p)
		for i := 0; i < s.Len(); i++ {
			ps = append(ps, s.Index(i).Interface())
		}
	} else {
		ps = append(ps, p.Interface())
	}
	return ps
}

func insert(t *testing.T, db DB, table string, records []Record) []Record {
	first := records[0]

	keys := make([]string, 0, len(first))
	columns := make([]string, 0, len(first))
	for column := range first {
		keys = append(keys, column)
		columns = append(columns, identifier(column))
	}

	params := make([]string, 0, len(records))
	values := make([]any, 0, len(records))
	param := 1

	for _, record := range records {
		tuple := make([]string, 0, len(first))
		for _, key := range keys {
			values = append(values, record[key])
			tuple = append(tuple, fmt.Sprintf("$%v", param))
			param++
		}
		params = append(params, fmt.Sprintf("(%v)", strings.Join(tuple, ", ")))
	}

	sql := fmt.Sprintf(`
		insert into %v (%v)
		values %v
		returning *
	`, identifier(table), strings.Join(columns, ", "), strings.Join(params, ", "))

	rows, err := db.Query(sql, values...)
	require.NoError(t, err, fmt.Sprintf("inserting row(s) into table %q", table))

	return fetchAll(t, rows)
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

func generate(retries int, f Factory, g []Generation, partials []any) ([]Record, map[int]string, error) {

	indexed := make(map[int]string)
	records := make([]Record, 0, len(partials))

EACH_PARTIAL:
	for _, p := range partials {
		attempts := 0

	EACH_RECORD:
		for {
			if attempts > retries {
				err := fmt.Errorf(
					"maximum %v retries exceeded generating record for table %q",
					retries,
					f.Table,
				)
				return nil, indexed, err
			}

			record := f.Generator()
			overrides := toRecord(p, strcase.ToSnake) // todo: recase function should be a config parameter
			override(record, overrides)

			for i, generation := range g {
				for j, uniqueBy := range f.UniqueBy {

					key := uniqueBy(record)
					if _, exists := generation[j][key]; exists {
						attempts++
						continue EACH_RECORD
					}
					if i == len(g)-1 {
						indexed[j] = key
						generation[j][key] = struct{}{}
					}
				}
			}

			records = append(records, record)

			continue EACH_PARTIAL
		}
	}
	return records, indexed, nil
}

func identifier(name string) string {
	parts := strings.Split(name, ".")
	for i, p := range parts {
		if !strings.HasPrefix(p, "\"") && !strings.HasSuffix(p, "\"") {
			parts[i] = "\"" + p + "\""
		}
	}
	return strings.Join(parts, ".")
}

func toRecord(s any, recase func(string) string) Record {
	v := reflect.Indirect(reflect.ValueOf(s))
	t := v.Type()
	r := make(Record, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		key := recase(t.Field(i).Name)
		value := v.Field(i)
		r[key] = value.Interface()
	}
	return r
}

func override(target Record, source Record) {
	for k := range target {
		if value, ok := source[k]; ok {
			target[k] = value
		}
	}
}
