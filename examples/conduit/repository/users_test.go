package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thebearingedge/dumbo"
	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres/schema"
	"github.com/thebearingedge/dumbo/examples/conduit/internal/conduittest"
)

func TestRegister(t *testing.T) {
	db := conduittest.RequireDB(t)
	conduittest.RequireTruncate(t, db, "users")

	t.Run("registers a user", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)

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
		tx := conduittest.RequireBegin(t, db)
		conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{"username": "gopher"})

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
		tx := conduittest.RequireBegin(t, db)
		conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{"email": "gopher@google.com"})

		users := NewUsersRepository(tx)

		created, err := users.Register(schema.Registration{
			Username: "gopher",
			Email:    "gopher@google.com",
			Password: "this should be hashed",
		})

		assert.NoError(t, err)
		assert.Nil(t, created)
	})
}

func TestFindByEmail(t *testing.T) {
	db := conduittest.RequireDB(t)
	tx := conduittest.RequireBegin(t, db)

	email := "gopher@google.com"
	conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{
		"email":    email,
		"username": "gopher",
	})

	t.Run("finds user with matching email", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		found, err := users.FindByEmail(email)

		assert.NoError(t, err)
		assert.Equal(t, "gopher", found.Username)
		assert.Equal(t, email, found.Email)
	})

	t.Run("does not find user with mismatched email", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		found, err := users.FindByEmail("rubyist@hey.com")

		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestFindByID(t *testing.T) {
	db := conduittest.RequireDB(t)
	tx := conduittest.RequireBegin(t, db)

	id := uint64(1)
	conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{
		"email":    "gopher@google.com",
		"username": "gopher",
	})

	t.Run("finds user with matching ID", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		found, err := users.FindByID(id)

		assert.NoError(t, err)
		assert.Equal(t, "gopher", found.Username)
		assert.Equal(t, "gopher@google.com", found.Email)
	})

	t.Run("does not find user with mismatched ID", func(t *testing.T) {
		sp := conduittest.RequireSavepoint(t, tx)

		users := NewUsersRepository(sp)

		found, err := users.FindByID(id + 1)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}
