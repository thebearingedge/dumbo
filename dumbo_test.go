package dumbo

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

type User struct {
	ID       sql.NullInt64
	Username sql.NullString
	Nickname string
	Age      sql.NullInt64
	IsSilly  sql.NullBool
}

func TestSeedingRecords(t *testing.T) {

	d := New(Schema{
		Table: "user",
		Columns: map[string]string{
			"id":       "ID",
			"username": "Username",
			"nickname": "Nickname",
			"age":      "Age",
			"is_silly": "IsSilly",
		},
	})

	db := dumbotest.RequireDB(t)

	t.Run("seeding a single record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		user := &User{
			Username: sql.NullString{String: "gopher", Valid: true},
			Age:      sql.NullInt64{Int64: 24, Valid: true},
			Nickname: "f",
		}

		d.Seed(t, tx, "user", user)

		assert.Equal(t, int64(1), user.ID.Int64)
		assert.Equal(t, "gopher", user.Username.String)
		assert.Equal(t, int64(24), user.Age.Int64)
		assert.False(t, user.IsSilly.Valid)
		assert.Equal(t, "f", user.Nickname)
	})

	t.Run("seeding multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		users := []*User{
			{
				Username: sql.NullString{String: "gopher", Valid: true},
				Age:      sql.NullInt64{Int64: 24, Valid: true},
			},
			{
				Username: sql.NullString{String: "rustacean", Valid: true},
				Age:      sql.NullInt64{Int64: 42, Valid: true},
				IsSilly:  sql.NullBool{Bool: true, Valid: true},
			},
		}

		d.Seed(t, tx, "user", users)

		gopher, rustacean := users[0], users[1]

		assert.Equal(t, int64(1), gopher.ID.Int64)
		assert.Equal(t, "gopher", gopher.Username.String)
		assert.Equal(t, int64(24), gopher.Age.Int64)
		assert.False(t, gopher.IsSilly.Valid)
		assert.Equal(t, "", gopher.Nickname)

		assert.Equal(t, int64(2), rustacean.ID.Int64)
		assert.Equal(t, "rustacean", rustacean.Username.String)
		assert.Equal(t, int64(42), rustacean.Age.Int64)
		assert.True(t, rustacean.IsSilly.Bool)
		assert.True(t, rustacean.IsSilly.Valid)
		assert.Equal(t, "", rustacean.Nickname)
	})
}
