package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type FavoritesRepository struct {
	db *postgres.DB
}

func NewFavoritesRepository(db *postgres.DB) FavoritesRepository {
	return FavoritesRepository{db: db}
}
