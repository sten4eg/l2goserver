package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"l2goserver/config"
	"runtime/trace"
)

var db *pgxpool.Pool

func ConfigureDB() error {
	conf := config.GetConfig()

	dsnString := "user=" + conf.LoginServer.Database.User
	dsnString += " password=" + conf.LoginServer.Database.Password
	dsnString += " host=" + conf.LoginServer.Database.Host
	dsnString += " port=" + conf.LoginServer.Database.Port
	dsnString += " dbname=" + conf.LoginServer.Database.Name
	dsnString += " sslmode=" + conf.LoginServer.Database.SSLMode
	dsnString += " search_path=" + conf.LoginServer.Database.Schema
	dsnString += " pool_max_conns=" + conf.LoginServer.Database.PoolMaxConn

	// todo лучше использовать unix socket,в БД выставить max_conn в районе 1000 и shared buffer ~ 1.5GB
	// unixWayPostgres := "postgresql:///postgres?host=/run/postgresql&port=5432&user=postgres&password=postgres&sslmode=disable"
	dbConfig, err := pgxpool.ParseConfig(dsnString)
	if err != nil {
		return err
	}

	// todo проверить simple и обычный
	dbConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return err
	}
	db = pool

	return nil
}

func GetConn() (conn *pgxpool.Conn, err error) {
	reg := trace.StartRegion(context.Background(), "DBAcquire")

	conn, err = db.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	reg.End()

	return
}
