package dumbotest

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func RequireDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	t.Cleanup(func() {
		if db == nil {
			return
		}
		if err := db.Close(); err != nil {
			panic(err)
		}
	})
	require.NoError(t, err)
	return db
}

func RequireBegin(t *testing.T, db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	t.Cleanup(func() {
		if tx == nil {
			return
		}
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	})
	require.NoError(t, err)
	return tx
}

func RequireSavepoint(t *testing.T, tx *sql.Tx) *sql.Tx {
	_, err := tx.Exec(fmt.Sprintf("savepoint %q", t.Name()))
	t.Cleanup(func() {
		if _, err := tx.Exec(fmt.Sprintf("rollback to savepoint %q", t.Name())); err != nil {
			panic(err)
		}
	})
	require.NoError(t, err)
	return tx
}
