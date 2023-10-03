package dumbo

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

func TestCreatingRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	seeder := New()

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

func TestFetchingRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	seeder := New()

	t.Run("fetching a single record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		seeder.SeedOne(t, tx, "user", Record{"username": "gopher"})
		user := seeder.FetchOne(t, tx, `select * from "user" where "username" = $1`, "gopher")

		assert.Equal(t, "gopher", user["username"])
	})

	t.Run("fetching multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		seeder.SeedMany(t, tx, "user", []Record{
			{"username": "gopher"},
			{"username": "rustacean"},
		})

		users := seeder.FetchMany(t, tx, `select * from "user" order by "id"`)

		gopher, rustacean := users[0], users[1]

		assert.Equal(t, int64(1), gopher["id"])
		assert.Equal(t, "gopher", gopher["username"])
		assert.Equal(t, int64(2), rustacean["id"])
		assert.Equal(t, "rustacean", rustacean["username"])
	})
}

func TestGeneratingRecordFields(t *testing.T) {
	db := dumbotest.RequireDB(t)

	seeder := New(
		Factory{
			Table: "user",
			NewRecord: func() Record {
				return Record{
					"username": faker.Username(),
				}
			},
		},
	)

	t.Run("generating one record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		random := seeder.SeedOne(t, tx, "user", Record{})

		assert.Equal(t, int64(1), random["id"])
		assert.NotEmpty(t, random["username"])
	})

	t.Run("generating multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		randoms := seeder.SeedMany(t, tx, "user", []Record{
			{},
			{},
			{},
		})

		r1, r2, r3 := randoms[0], randoms[1], randoms[2]

		assert.Equal(t, int64(1), r1["id"])
		assert.NotEmpty(t, r1["username"])
		assert.Equal(t, int64(2), r2["id"])
		assert.NotEmpty(t, r2["username"])
		assert.Equal(t, int64(3), r3["id"])
		assert.NotEmpty(t, r3["username"])
	})

	t.Run("overriding generated fields", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		gopher := seeder.SeedOne(t, tx, "user", Record{"username": "gopher"})

		assert.Equal(t, int64(1), gopher["id"])
		assert.Equal(t, "gopher", gopher["username"])
	})
}

func TestGeneratingUniqueRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	seeder := New(
		Factory{
			Table: "user",
			NewRecord: func() Record {
				return Record{
					"username": faker.Username(),
				}
			},
			UniqueBy: []Indexer{
				func(r Record) string {
					return fmt.Sprint(r["username"])
				},
			},
		},
	)

	t.Run("enforces unique seeds", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		assert.PanicsWithError(t, `maximum 5 retries exceeded generating record for table "user"`, func() {
			seeder.SeedMany(t, tx, "user", []Record{
				{"username": "gopher"},
				{"username": "gopher"},
			})
		})
	})

	t.Run("enforces unique inserts", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		assert.PanicsWithError(t, `maximum 5 retries exceeded generating record for table "user"`, func() {
			seeder.InsertMany(t, tx, "user", []Record{
				{"username": "gopher"},
				{"username": "gopher"},
			})
		})
	})
}

func TestNestedRuns(t *testing.T) {
	db := dumbotest.RequireDB(t)

	seeder := New(
		Factory{
			Table: "user",
			NewRecord: func() Record {
				return Record{
					"username": faker.Username(),
				}
			},
			UniqueBy: []Indexer{
				func(r Record) string {
					return fmt.Sprint(r["username"])
				},
			},
		},
	)

	t.Run("shared seed", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		seeder.SeedOne(t, tx, "user", Record{"username": "gopher"})

		t.Run("nested duplicate", func(t *testing.T) {
			sp := dumbotest.RequireSavepoint(t, tx)

			assert.PanicsWithError(t, `maximum 5 retries exceeded generating record for table "user"`, func() {
				seeder.Run(t, func(s *Dumbo) {
					s.InsertMany(t, sp, "user", []Record{
						{"username": "pythonista"},
						{"username": "gopher"},
					})
				})
			})
		})

		t.Run("nested unique", func(t *testing.T) {
			sp := dumbotest.RequireSavepoint(t, tx)

			assert.NotPanics(t, func() {
				seeder.Run(t, func(s *Dumbo) {
					s.InsertMany(t, sp, "user", []Record{
						{"username": "pythonista"},
						{"username": "rustacean"},
					})
				})
			})
		})
	})
}
