package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type followersRepository struct {
	db *postgres.DB
}

func NewFollowersRepository(db *postgres.DB) followersRepository {
	return followersRepository{db: db}
}
