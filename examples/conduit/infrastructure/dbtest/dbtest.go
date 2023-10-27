package dbtest

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/db"
)

type DB interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

type Tx struct {
	*db.Tx
}

func (t *Tx) Commit() error {
	_, err := t.Tx.Exec(`savepoint "committed"`)
	return err
}

func RequireDB(t *testing.T) *db.DB {
	t.Helper()

	t.Log(os.Getenv("DATABASE_URL"))
	pool, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, pool.Close()) })

	return &db.DB{DB: pool}
}

func RequireBegin(t *testing.T, db *db.DB) *Tx {
	t.Helper()

	tx, err := db.Begin()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, tx.Rollback()) })

	return &Tx{Tx: tx}
}

func RequireSavepoint(t *testing.T, tx *Tx) *Tx {
	t.Helper()

	_, err := tx.Exec(fmt.Sprintf("savepoint %q", t.Name()))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := tx.Exec(fmt.Sprintf("rollback to savepoint %q", t.Name()))
		require.NoError(t, err)
	})

	return tx
}

func RequireTruncate(t *testing.T, db DB, tables ...string) {
	t.Helper()

	names := make([]string, 0, len(tables))
	for _, name := range tables {
		names = append(names, fmt.Sprintf("%q", name))
	}

	_, err := db.Exec(fmt.Sprintf(`truncate table %v restart identity cascade`, strings.Join(names, ", ")))
	require.NoError(t, err)
}

func RequireScript(t *testing.T, db DB, scriptPath string) {
	t.Helper()

	_, caller, _, _ := runtime.Caller(1)
	relativePath := filepath.Join(filepath.Dir(caller), scriptPath)

	script, pathErr := filepath.Abs(relativePath)
	require.NoError(t, pathErr)

	sql, readErr := os.ReadFile(script)
	require.NoError(t, readErr)

	_, execErr := db.Exec(string(sql))
	require.NoError(t, execErr)
}

func RequireExec(t *testing.T, db DB, query string) {
	t.Helper()

	_, err := db.Exec(query)
	require.NoError(t, err)
}

func RequireRows(t *testing.T, db DB, query string) []map[string]any {
	t.Helper()

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	columns, err := rows.Columns()
	require.NoError(t, err)

	fetched := make([]map[string]any, 0)

	for rows.Next() {
		fields := make([]any, len(columns))
		for i := range fields {
			fields[i] = &fields[i]
		}

		require.NoError(t, rows.Scan(fields...))

		record := make(map[string]any, len(columns))
		for i, column := range columns {
			record[column] = fields[i]
		}

		fetched = append(fetched, record)
	}

	require.NoError(t, rows.Err())

	return fetched
}
