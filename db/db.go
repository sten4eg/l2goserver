package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"l2goserver/config"
)

var db *pgxpool.Pool

func ConfigureDB() {
	conf := config.GetConfig()

	dsnString := "user=" + conf.LoginServer.Database.User
	dsnString += " password=" + conf.LoginServer.Database.Password
	dsnString += " host=" + conf.LoginServer.Database.Host
	dsnString += " port=" + conf.LoginServer.Database.Port
	dsnString += " dbname=" + conf.LoginServer.Database.Name
	dsnString += " sslmode=" + conf.LoginServer.Database.SSLMode
	dsnString += " pool_max_conns=" + conf.LoginServer.Database.PoolMaxConn

	pool, err := pgxpool.Connect(context.Background(), dsnString)
	if err != nil {
		panic(err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		panic(err)
	}
	db = pool

}

func GetConn() (*pgxpool.Conn, error) {
	p, err := db.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	return p, nil
}
