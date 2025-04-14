package psql

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string, dbName string, migrationsDir string) (*Storage, error) {
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

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(data *models.UserData) error {
	return nil
}

func (s *Storage) SaveText(data string) error {
	return nil
}

func (s *Storage) GetText(id uuid.UUID) (*models.TextData, error) {
	return nil, nil
}

func (s *Storage) SaveBankCard(data *models.BankCardData) error {
	return nil
}

func (s *Storage) GetBankCard(id uuid.UUID) (*models.BankCardData, error) {
	return nil, nil
}

func (s *Storage) SaveMetaData(data *models.Meta) error {
	return nil
}

func (s *Storage) GetMetaData(id uuid.UUID) (*models.Meta, error) {
	return nil, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}
