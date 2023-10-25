package dumbo

import (
	"database/sql"
	"fmt"
	"reflect"
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

func (d Dumbo) Seed(t *testing.T, db DB, table string, partial any) {
	t.Helper()

	_, err := db.Exec(fmt.Sprintf(`truncate table %q restart identity cascade`, table))
	require.NoError(t, err, fmt.Sprintf("truncating table %q", table))

	d.Insert(t, db, table, partial)
}

func (d Dumbo) Insert(t *testing.T, db DB, table string, partial any) {

	partials := many(partial)
	schema := d.schemas[table]

	columns := make([]string, 0, len(schema.Columns))
	sFields := make([]string, 0, len(schema.Columns))
	for c, f := range schema.Columns {
		columns = append(columns, identifier(c))
		sFields = append(sFields, f)
	}

	params := make([]string, 0, len(partials)*len(columns))
	values := make([]any, 0, len(partials)*len(columns))
	inputs := make([][]any, 0, len(partials))
	outputs := make([][]any, 0, len(partials))
	param := 1

	for _, p := range partials {
		input := make([]any, 0, len(sFields))
		output := make([]any, 0, len(sFields))
		tuple := make([]string, 0, len(sFields))
		sValue := reflect.Indirect(reflect.ValueOf(p))
		for _, f := range sFields {
			var value any = sValue.FieldByName(f).Interface()
			switch value.(type) {
			case sql.NullString:
				if !value.(sql.NullString).Valid {
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
			case sql.NullBool:
				if !value.(sql.NullBool).Valid {
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
		inputs = append(inputs, input)
		outputs = append(outputs, output)
		values = append(values, input...)
		params = append(params, fmt.Sprintf("(%v)", strings.Join(tuple, ", ")))
	}

	query := fmt.Sprintf(`
		insert into %v (%v)
		values %v
		returning %v
	`, identifier(table), strings.Join(columns, ", "), strings.Join(params, ", "), strings.Join(columns, ", "))

	rows, err := db.Query(query, values...)
	require.NoError(t, err, fmt.Sprintf("inserting row(s) into table %v", identifier(table)))

	i := 0
	for rows.Next() {
		require.NoError(t, rows.Scan(outputs[i]...), "scanning row returned from query")
		i++
	}

	for i, p := range partials {
		record := reflect.Indirect(reflect.ValueOf(p))
		for j, f := range sFields {
			field := record.FieldByName(f)
			var value any
			v := reflect.Indirect(reflect.ValueOf(outputs[i][j]))
			switch field.Interface().(type) {
			case sql.NullString:
				var s string
				valid := !v.IsNil()
				if valid {
					s = v.Interface().(string)
				}
				value = sql.NullString{String: s, Valid: valid}
			case sql.NullInt64:
				var i int64
				valid := !v.IsNil()
				if valid {
					i = v.Interface().(int64)
				}
				value = sql.NullInt64{Int64: i, Valid: valid}
			case sql.NullBool:
				var b bool
				valid := !v.IsNil()
				if valid {
					b = v.Interface().(bool)
				}
				value = sql.NullBool{Bool: b, Valid: valid}
			}
			field.Set(reflect.ValueOf(value))
		}
	}

	require.NoError(t, rows.Err(), "iterating rows returned from query")
}

func many(partial any) []any {
	partials := make([]any, 0)
	partialValue := reflect.ValueOf(partial)
	if reflect.Indirect(partialValue).Kind() == reflect.Slice {
		slice := reflect.Indirect(partialValue)
		for i := 0; i < slice.Len(); i++ {
			partials = append(partials, slice.Index(i).Interface())
		}
	} else {
		partials = append(partials, partialValue.Interface())
	}
	return partials
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
