package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type TagsRepository struct {
	db *postgres.DB
}

func NewTagsRepository(db *postgres.DB) TagsRepository {
	return TagsRepository{db: db}
}
