package dumbotest

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq" // register postgres driver
	"github.com/stretchr/testify/require"
)

func RequireDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, err)
	return db
}

func RequireBegin(t *testing.T, db *sql.DB) *sql.Tx {
	t.Helper()
	tx, err := db.Begin()
	t.Cleanup(func() { require.NoError(t, tx.Rollback()) })
	require.NoError(t, err)
	return tx
}

func RequireSavepoint(t *testing.T, tx *sql.Tx) *sql.Tx {
	t.Helper()
	_, err := tx.Exec(fmt.Sprintf("savepoint %q", t.Name()))
	t.Cleanup(func() {
		_, err := tx.Exec(fmt.Sprintf("rollback to savepoint %q", t.Name()))
		require.NoError(t, err)
	})
	require.NoError(t, err)
	return tx
}
