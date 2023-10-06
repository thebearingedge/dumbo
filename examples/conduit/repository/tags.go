package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type tagsRepository struct {
	db *postgres.DB
}

func NewTagsRepository(db *postgres.DB) tagsRepository {
	return tagsRepository{db: db}
}
