package repository

import (
	"fmt"

	"github.com/thebearingedge/dumbo/examples/conduit/model"
)

type UsersRepository struct {
	db repository
}

func NewUsersRepository(db repository) UsersRepository {
	return UsersRepository{db: db}
}

func (r UsersRepository) Register(m model.Registration) (*model.Authentication, error) {
	sql := `
		insert into "users" ("email", "password", "username")
		values ($1, $2, $3)
		on conflict do nothing
		returning "id",
		          "email",
		          "username",
		          "bio",
		          "image_url"
	`

	rows, err := r.db.QueryRows(sql, m.Email, m.Password, m.Username)
	if err != nil {
		return nil, fmt.Errorf(`inserting into "users": %w`, err)
	}

	if !rows.Next() {
		return nil, nil
	}

	a := model.Authentication{}
	if err := rows.Scan(&a.ID, &a.Email, &a.Username, &a.Bio, &a.Image); err != nil {
		return nil, fmt.Errorf(`scanning from "users": %w`, err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating users "rows": %w`, err)
	}

	return &a, nil
}
