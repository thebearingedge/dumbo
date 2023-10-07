package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // register postgres driver
)

type DB struct {
	*sql.DB
}

func (d *DB) Begin() (*Tx, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}

	return &Tx{Tx: tx}, nil
}

type Tx struct {
	depth int
	*sql.Tx
}

func (t *Tx) Rollback() error {
	if t.depth == 0 {
		if err := t.Tx.Rollback(); err != nil {
			return fmt.Errorf("rolling back transaction: %w", err)
		}
		return nil
	}

	if _, err := t.Exec(fmt.Sprintf(`rollback to savepoint "savepoint_%v"`, t.depth)); err != nil {
		return fmt.Errorf(`rolling back to "savepoint_%v"`, t.depth)
	}

	return nil
}

func (t *Tx) QueryRows(sql string, values ...any) (*sql.Rows, error) {
	return t.Query(sql, values...)
}

func (t Tx) Begin() (*Tx, error) {
	_, err := t.Exec(fmt.Sprintf(`savepoint "savepoint_%v"`, t.depth))
	if err != nil {
		return nil, fmt.Errorf(`creating "savepoint_%v": %w`, t.depth, err)
	}

	return &Tx{Tx: t.Tx, depth: t.depth + 1}, nil
}
