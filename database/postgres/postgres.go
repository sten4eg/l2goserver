package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"l2goserver/config"
	"net/url"
	"time"
)

func NewPostgresDB(cfg config.DatabaseType) (*MainDB, error) {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable&search_path=%s",
		url.QueryEscape(cfg.User), url.QueryEscape(cfg.Password), cfg.Host,
		cfg.Port, cfg.Name, cfg.Schema)

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// для работы с PgBouncer
	conf.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	conf.MinConns = 5
	conf.HealthCheckPeriod = time.Minute * 5

	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	connector := stdlib.GetPoolConnector(pool)
	db := sql.OpenDB(connector)
	db.SetMaxIdleConns(0)

	m := &MainDB{
		baseDB: db,
	}

	return m, nil
}

type MainDB struct {
	baseDB *sql.DB
}

func (s *MainDB) Query(_ context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.baseDB.Query(query, args...)
}

func (s *MainDB) QueryRow(_ context.Context, query string, args ...any) *sql.Row {
	return s.baseDB.QueryRow(query, args...)
}

func (s *MainDB) Exec(_ context.Context, query string, args ...any) (sql.Result, error) {
	return s.baseDB.Exec(query, args...)
}

func (m *MainDB) Close() error {
	return m.baseDB.Close()
}
