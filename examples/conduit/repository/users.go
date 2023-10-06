package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type usersRepository struct {
	db *postgres.DB
}

func NewUsersRepository(db *postgres.DB) usersRepository {
	return usersRepository{db: db}
}
