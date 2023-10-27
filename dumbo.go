package dumbo

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type DB interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

func RequireDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, db.Close()) })

	return db
}

func RequireBegin(t *testing.T, db *sql.DB) *sql.Tx {
	t.Helper()

	tx, err := db.Begin()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, tx.Rollback()) })

	return tx
}

func RequireSavepoint(t *testing.T, tx *sql.Tx) *sql.Tx {
	t.Helper()

	_, err := tx.Exec(fmt.Sprintf("savepoint %q", t.Name()))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := tx.Exec(fmt.Sprintf("rollback to savepoint %q", t.Name()))
		require.NoError(t, err)
	})

	return tx
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
