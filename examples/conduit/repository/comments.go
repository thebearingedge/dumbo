package repository

import "github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres"

type CommentsRepository struct {
	db *postgres.DB
}

func NewCommentsRepository(db *postgres.DB) CommentsRepository {
	return CommentsRepository{db: db}
}
