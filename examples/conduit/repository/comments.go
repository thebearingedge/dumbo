package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type commentsRepository struct {
	db *postgres.DB
}

func NewCommentsRepository(db *postgres.DB) commentsRepository {
	return commentsRepository{db: db}
}
