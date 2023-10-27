package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/dbtest"

	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/db/schema"
)

func TestRegistration(t *testing.T) {
	db := dbtest.RequireDB(t)

	dbtest.RequireExec(t, db, `
		truncate table "users" restart identity cascade;
	`)

	t.Run("user registration", func(t *testing.T) {
		tx := dbtest.RequireBegin(t, db)

		users := NewUsersRepository(tx)

		created, err := users.Register(schema.Registration{
			Username: "gopher",
			Email:    "gopher@google.com",
			Password: "this should be hashed",
		})

		assert.NoError(t, err)
		assert.Equal(t, "gopher", created.Username)
		assert.Equal(t, "gopher@google.com", created.Email)
	})

	t.Run("does not register a duplicate username", func(t *testing.T) {
		tx := dbtest.RequireBegin(t, db)

		dbtest.RequireExec(t, tx, `
			insert into "users" ("username", "email", "password")
			values ('gopher', 'go@goo.com', 'this should be hashed');
		`)

		users := NewUsersRepository(tx)

		created, err := users.Register(schema.Registration{
			Username: "gopher",
			Email:    "gopher@google.com",
			Password: "this should be hashed",
		})

		assert.NoError(t, err)
		assert.Nil(t, created)
	})

	t.Run("does not register a duplicate email", func(t *testing.T) {
		tx := dbtest.RequireBegin(t, db)

		dbtest.RequireExec(t, tx, `
			insert into "users" ("username", "email", "password")
			values ('gopher', 'go@goo.com', 'this should be hashed');
		`)

		users := NewUsersRepository(tx)

		notCreated, err := users.Register(schema.Registration{
			Username: "go",
			Email:    "go@goo.com",
			Password: "this should be hashed",
		})

		assert.NoError(t, err)
		assert.Nil(t, notCreated)
	})
}

func TestFindByEmail(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	const email = "gopher@google.com"

	dbtest.RequireExec(t, tx, fmt.Sprintf(`
		truncate table "users" restart identity cascade;

		insert into "users" ("username", "email", "password")
		values ('gopher', '%s', 'this should be hashed');
	`, email))

	t.Run("finds user with matching email", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		found, err := users.FindByEmail(email)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "gopher", found.Username)
		assert.Equal(t, email, found.Email)
	})

	t.Run("does not find user with mismatched email", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		notFound, err := users.FindByEmail("rubyist@hey.com")

		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})
}

func TestFindByID(t *testing.T) {
	db := dbtest.RequireDB(t)
	tx := dbtest.RequireBegin(t, db)

	const id = uint64(1)

	dbtest.RequireExec(t, tx, `
		truncate table "users" restart identity cascade;

		insert into "users" ("username", "email", "password")
		values ('gopher', 'gopher@google.com', 'this should be hashed')
	`)

	t.Run("finds user with matching ID", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		found, err := users.FindByID(id)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "gopher", found.Username)
		assert.Equal(t, "gopher@google.com", found.Email)
	})

	t.Run("does not find user with mismatched ID", func(t *testing.T) {
		sp := dbtest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		notFound, err := users.FindByID(id + 1)

		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})
}