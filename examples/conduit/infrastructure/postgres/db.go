package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Transactor interface {
	Begin() (*Tx, error)
	Rollback() error
}

type DB struct {
	*sql.DB
	Transactor
}

func (d DB) Begin() (*Tx, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}

	return &Tx{Tx: tx, depth: 0}, nil
}

type Tx struct {
	depth uint
	*sql.Tx
	Transactor
}

func (t Tx) Rollback() error {
	if t.depth == 0 {
		return fmt.Errorf("rolling back transaction: %w", t.Tx.Rollback())
	}

	_, err := t.Tx.Exec(fmt.Sprintf(`rollback to savepoint "savepoint_%v"`, t.depth))
	if err != nil {
		return fmt.Errorf(`rolling back to "savepoint_%v"`, t.depth)
	}

	return nil
}

func (t Tx) Begin() (*Tx, error) {
	_, err := t.Tx.Exec(fmt.Sprintf(`savepoint "savepoint_%v"`, t.depth))
	if err != nil {
		return nil, fmt.Errorf(`creating "savepoint_%v": %w`, t.depth, err)
	}

	return &Tx{Tx: t.Tx, depth: t.depth + 1}, nil
}
