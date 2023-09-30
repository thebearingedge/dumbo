package dumbotest

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func RequireDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	})
	return db
}
