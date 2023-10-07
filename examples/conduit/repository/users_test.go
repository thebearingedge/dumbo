package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thebearingedge/dumbo"
	"github.com/thebearingedge/dumbo/examples/conduit/internal/conduittest"
	"github.com/thebearingedge/dumbo/examples/conduit/model"
)

func TestUsersRepository(t *testing.T) {
	db := conduittest.RequireDB(t)
	conduittest.RequireTruncate(t, db, "users")

	t.Run("registers a user", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		users := NewUsersRepository(tx)

		created, err := users.Register(model.Registration{
			Username: "gopher",
			Email:    "gopher@google.com",
			Password: "this should be hashed",
		})
		require.NoError(t, err)

		assert.Equal(t, "gopher", created.Username)
		assert.Equal(t, "gopher@google.com", created.Email)
	})

	t.Run("does not register a duplicate username", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{"username": "gopher"})
		users := NewUsersRepository(tx)

		created, err := users.Register(model.Registration{
			Username: "gopher",
			Email:    "gopher@google.com",
			Password: "this should be hashed",
		})
		require.NoError(t, err)

		assert.Nil(t, created)
	})

	t.Run("does not register a duplicate email", func(t *testing.T) {
		tx := conduittest.RequireBegin(t, db)
		conduittest.Seeder.SeedOne(t, tx, "users", dumbo.Record{"email": "gopher@google.com"})
		users := NewUsersRepository(tx)

		created, err := users.Register(model.Registration{
			Username: "gopher",
			Email:    "gopher@google.com",
			Password: "this should be hashed",
		})
		require.NoError(t, err)

		assert.Nil(t, created)
	})
}
