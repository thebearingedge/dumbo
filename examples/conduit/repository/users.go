package repository

import (
	"fmt"

	"github.com/thebearingedge/dumbo/examples/conduit/infrastructure/postgres/schema"
)

type UsersRepository struct {
	db repository
}

func NewUsersRepository(db repository) UsersRepository {
	return UsersRepository{db: db}
}

func (r UsersRepository) Register(m schema.Registration) (*schema.User, error) {
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
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	a := schema.User{}
	if err := rows.Scan(&a.ID, &a.Email, &a.Username, &a.Bio, &a.Image); err != nil {
		return nil, fmt.Errorf(`scanning from "users": %w`, err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating users "rows": %w`, err)
	}

	return &a, nil
}

func (r UsersRepository) FindByEmail(email string) (*schema.User, error) {
	sql := `
		select "id",
		       "email",
		       "username",
		       "bio",
		       "image_url"
		  from "users"
		 where "email" = $1
	`

	rows, err := r.db.QueryRows(sql, email)
	if err != nil {
		return nil, fmt.Errorf(`selecting from "users": %w`, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	a := schema.User{}
	if err := rows.Scan(&a.ID, &a.Email, &a.Username, &a.Bio, &a.Image); err != nil {
		return nil, fmt.Errorf(`scanning from "users": %w`, err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating "users" rows: %w`, err)
	}

	return &a, nil
}

func (r UsersRepository) FindByID(id uint64) (*schema.User, error) {
	sql := `
		select "id",
		       "email",
		       "username",
		       "bio",
		       "image_url"
		  from "users"
		 where "id" = $1
	`

	rows, err := r.db.QueryRows(sql, id)
	if err != nil {
		return nil, fmt.Errorf(`selecting from "users": %w`, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	a := schema.User{}
	if err := rows.Scan(&a.ID, &a.Email, &a.Username, &a.Bio, &a.Image); err != nil {
		return nil, fmt.Errorf(`scanning from "users": %w`, err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`iterating "users" rows: %w`, err)
	}

	return &a, nil
}
