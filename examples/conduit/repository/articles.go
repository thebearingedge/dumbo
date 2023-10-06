package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type articlesRepository struct {
	db *postgres.DB
}

func NewArticlesRepository(db *postgres.DB) articlesRepository {
	return articlesRepository{db: db}
}
