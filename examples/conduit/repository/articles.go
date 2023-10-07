package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type ArticlesRepository struct {
	db *postgres.DB
}

func NewArticlesRepository(db *postgres.DB) ArticlesRepository {
	return ArticlesRepository{db: db}
}
