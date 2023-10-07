package conduittest

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/thebearingedge/dumbo"
	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"
)

func RequireDB(t *testing.T) *postgres.DB {
	t.Helper()
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, err)
	return &postgres.DB{DB: db}
}

func RequireBegin(t *testing.T, db *postgres.DB) *postgres.Tx {
	t.Helper()
	tx, err := db.Begin()
	t.Cleanup(func() { require.NoError(t, tx.Rollback()) })
	require.NoError(t, err)
	return tx
}

func RequireTruncate(t *testing.T, db *postgres.DB, tables ...string) {
	t.Helper()
	names := make([]string, 0, len(tables))
	for _, name := range tables {
		names = append(names, fmt.Sprintf("%q", name))
	}
	_, err := db.Exec(fmt.Sprintf(`truncate table %v restart identity cascade`, strings.Join(names, ", ")))
	require.NoError(t, err)
}

func RequireSavepoint(t *testing.T, tx *postgres.Tx) *postgres.Tx {
	t.Helper()
	_, err := tx.Exec(fmt.Sprintf("savepoint %q", t.Name()))
	t.Cleanup(func() {
		_, err := tx.Exec(fmt.Sprintf("rollback to savepoint %q", t.Name()))
		require.NoError(t, err)
	})
	require.NoError(t, err)
	return tx
}

var Seeder dumbo.Dumbo = dumbo.New(
	dumbo.Factory{
		Table: "users",
		NewRecord: func() dumbo.Record {
			return dumbo.Record{
				"username": faker.Username(),
				"email":    faker.Email(),
				"password": faker.Password(),
			}
		},
		UniqueBy: []dumbo.Indexer{
			func(r dumbo.Record) string {
				return fmt.Sprint(r["username"])
			},
			func(r dumbo.Record) string {
				return fmt.Sprint(r["email"])
			},
		},
	},
)
