package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type favoritesRepository struct {
	db *postgres.DB
}

func NewFavoritesRepository(db *postgres.DB) favoritesRepository {
	return favoritesRepository{db: db}
}
