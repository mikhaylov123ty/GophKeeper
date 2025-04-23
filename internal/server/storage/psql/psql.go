package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

func (s *Storage) SaveText(data *models.TextData) error {
	query, args, err := squirrel.Insert(textTableName).
		Values(data.ID, data.Text).
		Suffix("ON CONFLICT(id) DO UPDATE SET text = $2").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save text query: %w", err)
	}

	slog.Debug("saving text data", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not save text data: %w", err)
	}

	return nil
}

func (s *Storage) GetTextByID(id uuid.UUID) (*models.TextData, error) {
	query, args, err := squirrel.Select("*").
		From(textTableName).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not build get text by id query: %w", err)
	}

	slog.Debug("getting meta data", slog.String("query", query), slog.Any("args", args))

	row := s.db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, fmt.Errorf("could not execute get text by id query: %w", row.Err())
	}
	var res models.TextData
	if err = row.Scan(
		&res.ID,
		&res.Text,
	); err != nil {
		return nil, fmt.Errorf("could not scan get text by id query: %w", err)
	}

	return &res, nil
}

func (s *Storage) SaveBankCard(data *models.BankCardData) error {
	query, args, err := squirrel.Insert(bankCardsTableName).
		Values(data.ID, data.CardNum, data.Expiry, data.CVV).
		Suffix("ON CONFLICT(id) DO UPDATE SET card_num = $2, expiry = $3, cvv = $4").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save bank card query: %w", err)
	}

	slog.Debug("saving bank card data", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not save bank card data: %w", err)
	}

	return nil
}

func (s *Storage) GetBankCardById(id uuid.UUID) (*models.BankCardData, error) {
	query, args, err := squirrel.Select("*").
		From(bankCardsTableName).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not build get bank card query: %w", err)
	}

	slog.Debug("get bank card data", slog.String("query", query), slog.Any("args", args))

	row := s.db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, fmt.Errorf("could not execute get bank card query: %w", row.Err())
	}

	var res models.BankCardData
	if err = row.Scan(
		&res.ID,
		&res.CardNum,
		&res.Expiry,
		&res.CVV,
	); err != nil {
		return nil, fmt.Errorf("could not scan get bank card query row to struct: %w", row.Err())
	}

	return &res, nil
}

func (s *Storage) SaveMetaData(data *models.Meta) error {
	query, args, err := squirrel.Insert(metaTableName).
		Values(data.ID, data.Title, data.Description, data.Type, data.DataID, data.UserID, data.Created, data.Modified).
		Suffix("ON CONFLICT(id) DO UPDATE SET title = $2, description = $3, data_id = $5, modified_at = $7").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save meta query: %w", err)
	}

	slog.Debug("saving meta data", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not execute save meta query: %w", err)
	}

	return nil
}

func (s *Storage) GetMetaDataByUser(userID uuid.UUID) ([]*models.Meta, error) {
	query, args, err := squirrel.Select("*").
		From(metaTableName).
		Where(
			squirrel.And{
				squirrel.Eq{"user_id": userID},
			}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not build get meta query: %w", err)
	}

	slog.Debug("getting meta data", slog.String("query", query), slog.Any("args", args))

	rows, err := s.db.Query(query, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("could not execute get meta query: %w", err)
	}
	defer rows.Close()

	var res []*models.Meta
	for rows.Next() {
		row := &models.Meta{}
		if err = rows.Scan(
			&row.ID,
			&row.Title,
			&row.Description,
			&row.Type,
			&row.DataID,
			&row.UserID,
			&row.Modified,
			&row.Created,
		); err != nil {
			return nil, fmt.Errorf("could not execute get meta query: %w", err)
		}

		res = append(res, row)
	}

	if len(res) == 0 {
		return nil, sql.ErrNoRows
	}

	return res, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}
