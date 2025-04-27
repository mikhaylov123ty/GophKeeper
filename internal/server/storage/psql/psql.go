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
	"log/slog"
)

const (
	metaTableName      = "metas"
	itemsDataTableName = "items_data"
	usersTableName     = "users"
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
	query, args, err := squirrel.Insert(usersTableName).
		Values(data.ID, data.Login, data.Password, data.Created, data.Modified).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save user query: %w", err)
	}

	slog.Debug("saving user", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not save user: %w", err)
	}

	return nil
}

func (s *Storage) GetUserByLogin(login string) (*models.UserData, error) {
	query, args, err := squirrel.Select("*").
		From(usersTableName).
		Where(squirrel.Eq{"login": login}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not build get user query: %w", err)
	}

	slog.Debug("getting user query", slog.String("query", query), slog.Any("args", args))

	row := s.db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, fmt.Errorf("could not execute get user query: %w", row.Err())
	}

	var user models.UserData
	if err = row.Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Created,
		&user.Modified,
	); err != nil {
		return nil, fmt.Errorf("could not scan get user data: %w", err)
	}

	return &user, nil
}

func (s *Storage) SaveItemData(item *models.ItemData) error {
	query, args, err := squirrel.Insert(itemsDataTableName).
		Values(item.ID, item.Data).
		Suffix("ON CONFLICT(id) DO UPDATE SET data = $2").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save item data query: %w", err)
	}

	slog.Debug("saving item data", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not save item data: %w", err)
	}

	return nil
}

func (s *Storage) GetItemDataByID(id uuid.UUID) (*models.ItemData, error) {
	query, args, err := squirrel.Select("*").
		From(itemsDataTableName).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not build get item data by id query: %w", err)
	}

	slog.Debug("getting data", slog.String("query", query), slog.Any("args", args))

	row := s.db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, fmt.Errorf("could not execute get item data by id query: %w", row.Err())
	}
	var res models.ItemData
	if err = row.Scan(
		&res.ID,
		&res.Data,
	); err != nil {
		return nil, fmt.Errorf("could not scan get item data by id query: %w", err)
	}

	return &res, nil
}

func (s *Storage) DeleteItemDataByID(id uuid.UUID) error {
	query, args, err := squirrel.Delete(itemsDataTableName).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build delete item data by id query: %w", err)
	}

	slog.Debug("deleting item data", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not delete item data: %w", err)
	}

	return nil
}

func (s *Storage) SaveMetaData(data *models.Meta) error {
	fmt.Printf("META: %+v", *data)
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

func (s *Storage) DeleteMetaDataById(id uuid.UUID) error {
	query, args, err := squirrel.Delete(metaTableName).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build delete meta by id query: %w", err)
	}

	slog.Debug("deleting metadata", slog.String("query", query), slog.Any("args", args))

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not delete meta data: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}
