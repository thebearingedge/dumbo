package conduittest

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"
)

func RequireDB(t *testing.T) *postgres.DB {
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
	return &postgres.DB{DB: db}
}

func RequireBegin(t *testing.T, db postgres.Transactor) *postgres.Tx {
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
