package dumbo

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type DB interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

type Schema struct {
	Table   string
	Columns map[string]string
}

type Dumbo struct {
	schemas map[string]Schema
}

func New(schemas ...Schema) Dumbo {
	s := make(map[string]Schema, len(schemas))
	for _, schema := range schemas {
		s[schema.Table] = schema
	}
	return Dumbo{
		schemas: s,
	}
}

func (d Dumbo) SeedFile(t *testing.T, db DB, path string) {
	t.Helper()

	_, caller, _, _ := runtime.Caller(1)
	script, pathErr := filepath.Abs(filepath.Join(filepath.Dir(caller), path))
	require.NoError(t, pathErr)

	sql, readErr := os.ReadFile(script)
	require.NoError(t, readErr)

	_, execErr := db.Exec(string(sql))
	require.NoError(t, execErr)
}

func (d Dumbo) Seed(t *testing.T, db DB, table string, record any) {
	t.Helper()

	_, err := db.Exec(fmt.Sprintf(`truncate table %v restart identity cascade`, identifier(table)))
	require.NoError(t, err, fmt.Sprintf("truncating table %v", identifier(table)))

	d.Insert(t, db, table, record)
}

func (d Dumbo) Insert(t *testing.T, db DB, table string, record any) {
	t.Helper()

	records := many(record)
	schema := d.schemas[table]

	columns := make([]string, 0, len(schema.Columns))
	sFields := make([]string, 0, len(schema.Columns))
	for c, f := range schema.Columns {
		columns = append(columns, identifier(c))
		sFields = append(sFields, f)
	}

	tuples := make([]string, 0, len(records))
	inputs := make([]any, 0, len(records)*len(columns))
	param := 1

	for _, p := range records {
		input := make([]any, 0, len(sFields))
		output := make([]any, 0, len(sFields))
		tuple := make([]string, 0, len(sFields))
		sValue := reflect.Indirect(reflect.ValueOf(p))
		for _, f := range sFields {
			var value any = sValue.FieldByName(f).Interface()
			switch value.(type) {
			case sql.NullBool:
				if !value.(sql.NullBool).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullByte:
				if !value.(sql.NullByte).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullFloat64:
				if !value.(sql.NullFloat64).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullInt16:
				if !value.(sql.NullInt16).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullInt32:
				if !value.(sql.NullInt32).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullInt64:
				if !value.(sql.NullInt64).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullString:
				if !value.(sql.NullString).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			case sql.NullTime:
				if !value.(sql.NullTime).Valid {
					tuple = append(tuple, "default")
				} else {
					input = append(input, value)
					tuple = append(tuple, fmt.Sprintf("$%d", param))
					param++
				}
			default:
				input = append(input, value)
				tuple = append(tuple, fmt.Sprintf("$%d", param))
				param++
			}
			output = append(output, &value)
		}
		inputs = append(inputs, input...)
		tuples = append(tuples, fmt.Sprintf("(%v)", strings.Join(tuple, ", ")))
	}

	query := fmt.Sprintf(`
		insert into %v (%v)
		values %v
		returning %v
	`, identifier(table), strings.Join(columns, ", "), strings.Join(tuples, ", "), strings.Join(columns, ", "))

	rows, err := db.Query(query, inputs...)
	require.NoError(t, err, fmt.Sprintf("inserting row(s) into table %v with query: \n%v\n", identifier(table), query))

	i := 0
	for rows.Next() {
		outputs := make([]any, 0, len(columns))
		sValue := reflect.ValueOf(records[i]).Elem()
		for _, sField := range sFields {
			sField := sValue.FieldByName(sField).Addr().Interface()
			outputs = append(outputs, sField)
		}
		require.NoError(t, rows.Scan(outputs...), "scanning row returned from query")
		i++
	}

	require.NoError(t, rows.Err(), "iterating rows returned from query")
}

func many(record any) []any {
	value := reflect.ValueOf(record)
	if reflect.Indirect(value).Kind() != reflect.Slice {
		return []any{value.Interface()}
	}
	slice := reflect.Indirect(value)
	records := make([]any, 0, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		records = append(records, slice.Index(i).Interface())
	}
	return records
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
