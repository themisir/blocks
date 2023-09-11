package main

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"log"
)

//go:embed migrations
var migrations embed.FS

func CreateMigratedDb(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %e", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migration driver: %e", err)
	}

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migration source: %e", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrations: %e", err)
	}

	if err := m.Up(); err != nil {
		log.Println(err)
	}

	return db, nil
}
