package dumbo

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID       sql.NullInt64
	Username sql.NullString
	Nickname string
	Age      sql.NullInt64
	IsSilly  sql.NullBool
}

func TestSeedingScripts(t *testing.T) {

	db := RequireDB(t)

	t.Run("seeding from a script file", func(t *testing.T) {
		tx := RequireBegin(t, db)
		RequireExec(t, tx, `--sql
			truncate table "user" restart identity cascade;
			insert into "user" (
				"username",
				"nickname",
				"age",
				"is_silly"
			)
			values
				('gopher', 'nibbles', 24, default),
				('rustacean', 'crab', 42, true);
		`)

		query := `--sql
			select
				"id",
				"username",
				"nickname",
				"age",
				"is_silly"
			from "user"
			order by "id"
		`

		rows, err := tx.Query(query)
		assert.NoError(t, err)

		users := make([]User, 0, 2)
		for rows.Next() {
			u := User{}
			err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.Age, &u.IsSilly)
			assert.NoError(t, err)
			users = append(users, u)
		}

		gopher, rustacean := users[0], users[1]

		assert.Equal(t, int64(1), gopher.ID.Int64)
		assert.Equal(t, "gopher", gopher.Username.String)
		assert.Equal(t, "nibbles", gopher.Nickname)
		assert.Equal(t, int64(24), gopher.Age.Int64)
		assert.False(t, gopher.IsSilly.Valid)

		assert.Equal(t, int64(2), rustacean.ID.Int64)
		assert.Equal(t, "rustacean", rustacean.Username.String)
		assert.Equal(t, int64(42), rustacean.Age.Int64)
		assert.Equal(t, "crab", rustacean.Nickname)
		assert.True(t, rustacean.IsSilly.Bool)
		assert.True(t, rustacean.IsSilly.Valid)
	})
}
