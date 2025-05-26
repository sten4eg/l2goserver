package main

import (
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"l2goserver/config"
	"l2goserver/migrations"
	"log"
	"net/url"
	"strconv"
)

func main() {

	err := config.Read()
	cfg := config.GetConfig()
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(cfg.LoginServer.Database.Port)
	if err != nil {
		panic(err)
	}
	err = Migrate(Config{
		Credentials: Credentials{
			Username: cfg.LoginServer.Database.User,
			Password: cfg.LoginServer.Database.Password,
			Host:     cfg.LoginServer.Database.Host,
			Port:     port,
			Database: cfg.LoginServer.Database.Schema,
		},
		Fs:     migrations.FS,
		FsPath: ".",
	})
	if err != nil {
		log.Fatal(err)
	}
}

type (
	// Credentials - данные необходимые для выполнения миграции
	Credentials struct {
		Username string
		Password string
		Host     string
		Port     int
		Database string
		Scheme   string
	}
)

type (
	// Config конфигурация для миграций при чтении из встроенной файловой системы
	Config struct {
		Credentials
		Fs     embed.FS
		FsPath string
	}
)

// Migrate - выполнение миграции
func Migrate(config Config) error {
	driver, err := iofs.New(config.Fs, config.FsPath)
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		url.QueryEscape(config.Username), url.QueryEscape(config.Password), config.Host, config.Port, config.Database, config.Scheme)

	m, err := migrate.NewWithSourceInstance("iofs", driver, dsn)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
