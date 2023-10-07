package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type FollowersRepository struct {
	db *postgres.DB
}

func NewFollowersRepository(db *postgres.DB) FollowersRepository {
	return FollowersRepository{db: db}
}
