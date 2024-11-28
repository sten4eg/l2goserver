package database

import (
	"context"
	"database/sql"
	"l2goserver/config"
	"l2goserver/database/postgres"
)

type Database interface {
	Query(context.Context, string, ...any) (*sql.Rows, error)
	QueryRow(context.Context, string, ...any) *sql.Row
	Exec(context.Context, string, ...any) (sql.Result, error)
	Close() error
}

func NewDatabase(conf config.DatabaseType) (Database, error) {
	dbs, err := postgres.NewPostgresDB(conf)
	if err != nil {
		return nil, err
	}
	return dbs, nil
}
