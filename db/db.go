package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"l2goserver/config"
)

var db *pgxpool.Pool

var dbGS []*pgxpool.Pool

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

	configureDbToGameServers(conf)
}

func configureDbToGameServers(conf config.Conf) {
	for _, gameServer := range conf.GameServers {

		dsnString := "user=" + gameServer.Database.User
		dsnString += " password=" + gameServer.Database.Password
		dsnString += " host=" + gameServer.Database.Host
		dsnString += " port=" + gameServer.Database.Port
		dsnString += " dbname=" + gameServer.Database.Name
		dsnString += " sslmode=" + gameServer.Database.SSLMode
		dsnString += " pool_max_conns=" + gameServer.Database.PoolMaxConn
		pool, err := pgxpool.Connect(context.Background(), dsnString)
		if err != nil {
			panic(err)
		}
		err = pool.Ping(context.Background())
		if err != nil {
			panic(err)
		}
		dbGS = append(dbGS, pool)
	}
}
func GetConn() (*pgxpool.Conn, error) {
	p, err := db.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	return p, nil
}
func CountDBConn() int {
	return len(dbGS)
}

func GetConnToGS(index int) (*pgxpool.Conn, error) {
	if index > CountDBConn() || index < 0 {
		return nil, errors.New("такого нет")
	}
	p, err := dbGS[index].Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	return p, nil
}
