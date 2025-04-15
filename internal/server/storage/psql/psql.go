package psql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
)

const (
	metaTableName      = "metas"
	textTableName      = "texts"
	usersTableName     = "users"
	bankCardsTableName = "bank_cards"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string, dbName string, migrationsDir string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
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

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
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
	query, args, err := squirrel.Insert(metaTableName).Values(data.ID, data.Title, data.Description,
		data.Type, data.DataID, data.Created, data.Modified).
		Suffix("ON CONFLICT(id) DO UPDATE SET title = $2, description = $3, data_id = $5, modified_at = $7").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save meta query: %w", err)
	}

	fmt.Println("QUERY", query)
	fmt.Println("ARGS", args)

	res, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not execute save meta query: %w", err)
	}
	i, err := res.RowsAffected()

	fmt.Println("res", i)

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
