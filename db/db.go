package db

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"l2goserver/config"
	"strconv"
)

func ConfigureDB() (*sql.DB, error) {
	conf := config.GetConfig()

	dsnString := "user=" + conf.LoginServer.Database.User
	dsnString += " password=" + conf.LoginServer.Database.Password
	dsnString += " host=" + conf.LoginServer.Database.Host
	dsnString += " port=" + conf.LoginServer.Database.Port
	dsnString += " dbname=" + conf.LoginServer.Database.Name
	dsnString += " sslmode=" + conf.LoginServer.Database.SSLMode
	dsnString += " search_path=" + conf.LoginServer.Database.Schema
	dsnString += " pool_max_conns=" + conf.LoginServer.Database.PoolMaxConn

	// unixWayPostgres := "postgresql:///postgres?host=/run/postgresql&port=5432&user=postgres&password=postgres&sslmode=disable"
	dbConfig, err := pgxpool.ParseConfig(dsnString)
	if err != nil {
		return nil, err
	}

	dbConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)
	maxConni, err := strconv.Atoi(conf.LoginServer.Database.PoolMaxConn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConni)
	return db, nil

}
