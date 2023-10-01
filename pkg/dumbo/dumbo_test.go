package dumbo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

func TestCreatingRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	seeder := NewSeeder()

	t.Run("seeding a single record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		user := seeder.SeedOne(t, tx, "user", Record{"username": "gopher"})

		assert.Equal(t, "gopher", user["username"])
	})

	t.Run("seeding multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		users := seeder.SeedMany(t, tx, "user", []Record{
			{"username": "gopher"},
			{"username": "rustacean"},
		})

		gopher, rustacean := users[0], users[1]

		assert.Equal(t, int64(1), gopher["id"])
		assert.Equal(t, "gopher", gopher["username"])
		assert.Equal(t, int64(2), rustacean["id"])
		assert.Equal(t, "rustacean", rustacean["username"])
	})

	t.Run("adding a single record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		gopher := seeder.SeedOne(t, tx, "user", Record{"username": "gopher"})

		assert.Equal(t, int64(1), gopher["id"])
		assert.Equal(t, "gopher", gopher["username"])

		rustacean := seeder.InsertOne(t, tx, "user", Record{"username": "rustacean"})

		assert.Equal(t, int64(2), rustacean["id"])
		assert.Equal(t, "rustacean", rustacean["username"])
	})

	t.Run("adding multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		gopher := seeder.SeedOne(t, tx, "user", Record{"username": "gopher"})

		assert.Equal(t, int64(1), gopher["id"])
		assert.Equal(t, "gopher", gopher["username"])

		mls := seeder.InsertMany(t, tx, "user", []Record{
			{"username": "rust"},
			{"username": "ocaml"},
		})

		rust, ocaml := mls[0], mls[1]

		assert.Equal(t, int64(2), rust["id"])
		assert.Equal(t, "rust", rust["username"])

		assert.Equal(t, int64(3), ocaml["id"])
		assert.Equal(t, "ocaml", ocaml["username"])
	})
}
