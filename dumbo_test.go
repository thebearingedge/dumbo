package dumbo

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/internal/dumbotest"
)

type User struct {
	ID       int64
	Username string
}

func TestAddingRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	d := New(Factory{
		Table: "user",
		Generator: func() Record {
			return Record{
				"username": faker.Username(),
			}
		},
	})

	t.Run("seeding a single record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		user := &User{
			Username: "gopher",
		}

		d.Seed(t, tx, "user", user)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "gopher", user.Username)
	})

	t.Run("seeding multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		users := []*User{
			{Username: "gopher"},
			{Username: "rustacean"},
		}

		d.Seed(t, tx, "user", users)

		gopher, rustacean := users[0], users[1]

		assert.Equal(t, int64(1), gopher.ID)
		assert.Equal(t, "gopher", gopher.Username)
		assert.Equal(t, int64(2), rustacean.ID)
		assert.Equal(t, "rustacean", rustacean.Username)
	})

	t.Run("adding a single record", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		d.Seed(t, tx, "user", &User{
			Username: "gopher",
		})

		rustacean := &User{
			Username: "rustacean",
		}

		d.Insert(t, tx, "user", rustacean)

		assert.Equal(t, int64(2), rustacean.ID)
		assert.Equal(t, "rustacean", rustacean.Username)
	})

	t.Run("adding multiple records", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		d.Seed(t, tx, "user", &User{
			Username: "gopher",
		})

		rest := []*User{
			{Username: "rustacean"},
			{Username: "pythonista"},
		}

		d.Insert(t, tx, "user", rest)

		rustacean, pythonista := rest[0], rest[1]

		assert.Equal(t, int64(2), rustacean.ID)
		assert.Equal(t, "rustacean", rustacean.Username)
		assert.Equal(t, int64(3), pythonista.ID)
		assert.Equal(t, "pythonista", pythonista.Username)
	})
}

func TestUniqueRecords(t *testing.T) {
	db := dumbotest.RequireDB(t)

	d := New(Factory{
		Table: "user",
		Generator: func() Record {
			return Record{
				"username": faker.Username(),
			}
		},
		UniqueBy: []Indexer{
			func(r Record) string {
				return fmt.Sprint(r["username"])
			},
		},
	})

	t.Run("enforces unique seeds", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		assert.PanicsWithError(t, `generating records: maximum 5 retries exceeded generating record for table "user"`, func() {
			d.Seed(t, tx, "user", []*User{
				{Username: "gopher"},
				{Username: "gopher"},
			})
		})
	})

	t.Run("enforces unique inserts", func(t *testing.T) {
		tx := dumbotest.RequireBegin(t, db)

		assert.PanicsWithError(t, `generating records: maximum 5 retries exceeded generating record for table "user"`, func() {
			d.Insert(t, tx, "user", []*User{
				{Username: "gopher"},
				{Username: "gopher"},
			})
		})
	})
}

func TestNestedGenerations(t *testing.T) {
	db := dumbotest.RequireDB(t)

	d := New(
		Factory{
			Table: "user",
			Generator: func() Record {
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

		d.Seed(t, tx, "user", &User{
			Username: "gopher",
		})

		t.Run("nested duplicate", func(t *testing.T) {
			sp := dumbotest.RequireSavepoint(t, tx)

			assert.PanicsWithError(t, `generating records: maximum 5 retries exceeded generating record for table "user"`, func() {
				d.Insert(t, sp, "user", []*User{
					{Username: "pythonista"},
					{Username: "gopher"},
				})
			})
		})

		t.Run("nested unique", func(t *testing.T) {
			sp := dumbotest.RequireSavepoint(t, tx)

			assert.NotPanics(t, func() {
				d.Insert(t, sp, "user", []*User{
					{Username: "pythonista"},
					{Username: "rustacean"},
				})
			})
		})
	})
}
