package psql

import (
	"database/sql"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	migrationsDir = "file://./internal/server/storage/migrations"
)

type DB struct {
	conn *sql.DB
}

func New(dsn string, dbName string, migrationsDir string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not open postgres database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not init postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		dbName, driver)
	if err != nil {
		return nil, fmt.Errorf("could not init postgres migrate: %w", err)
	}

	if err = m.Up(); err != nil {
		return nil, fmt.Errorf("could not apply migrations: %w", err)
	}

	return &DB{conn: db}, nil
}

func (db *DB) SaveUser(data *models.UserData) error {
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
