package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mikhaylov123ty/GophKeeper/internal/domain"
)

const (
	metaTableName      = "metas"
	itemsDataTableName = "items_data"
	usersTableName     = "users"
)

// Storage represents a storage layer that handles database operations using an SQL database connection.
type Storage struct {
	db *sql.DB
}

func New(dsn string, migrationsDir string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not open postgres database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not init postgres driver: %w", err)
	}

	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not parse dsn: %w", err)
	}

	dbName, ok := strings.CutPrefix(parsedDSN.Path, "/")
	if !ok {
		return nil, fmt.Errorf("could not cut prefix database name")
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

// SaveUser inserts a new user record into the database or returns an error if the operation fails.
func (s *Storage) SaveUser(data *domain.UserData) error {
	slog.Debug("Save User Data", slog.Any("data", *data))

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

// GetUserByLogin retrieves a user by their login from the database and returns the corresponding UserData or an error.
func (s *Storage) GetUserByLogin(login string) (*domain.UserData, error) {
	slog.Debug("Get User Data by Login", slog.String("Login", login))

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

	var user domain.UserData
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

// SaveItemData saves the provided item data and its associated metadata in the database within a transaction.
// The method updates existing records or inserts new ones based on ID conflicts, and returns an error if any step fails.
func (s *Storage) SaveItemData(item *domain.ItemData, meta *domain.Meta) error {
	slog.Debug("Save Item Data", slog.Any("data", *item))
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	itemDataQuery, itemDataArgs, err := squirrel.Insert(itemsDataTableName).
		Values(item.ID, item.Data).
		Suffix("ON CONFLICT(id) DO UPDATE SET data = $2").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save item data query: %w", err)
	}

	metaDataQuery, metaDataArgs, err := squirrel.Insert(metaTableName).
		Values(meta.ID, meta.Title, meta.Description, meta.Type, meta.DataID, meta.UserID, meta.Created, meta.Modified).
		Suffix("ON CONFLICT(id) DO UPDATE SET title = $2, description = $3, data_id = $5, modified_at = $7").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build save meta query: %w", err)
	}

	slog.Debug("saving item data", slog.String("query", itemDataQuery), slog.Any("args", itemDataArgs))

	_, err = tx.Exec(itemDataQuery, itemDataArgs...)
	if err != nil {
		return fmt.Errorf("could not save item data: %w", err)
	}

	slog.Debug("saving meta data", slog.String("query", metaDataQuery), slog.Any("args", metaDataArgs))

	_, err = tx.Exec(metaDataQuery, metaDataArgs...)
	if err != nil {
		return fmt.Errorf("could not save meta data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// GetItemDataByID retrieves the item data by its unique ID from the items_data table and returns it or an error.
func (s *Storage) GetItemDataByID(id uuid.UUID) (*domain.ItemData, error) {
	slog.Debug("Get Item Data by ID", slog.String("ID", id.String()))
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
	var res domain.ItemData
	if err = row.Scan(
		&res.ID,
		&res.Data,
	); err != nil {
		return nil, fmt.Errorf("could not scan get item data by id query: %w", err)
	}

	return &res, nil
}

// DeleteItemDataByID removes an item record from the items_data table based on its unique ID. Returns an error if it fails.
func (s *Storage) DeleteItemDataByID(id uuid.UUID) error {
	slog.Debug("Delete Item Data by ID", slog.String("ID", id.String()))

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

// GetMetaDataByUser retrieves metadata records associated with a specific user ID from the database or returns an error.
func (s *Storage) GetMetaDataByUser(userID uuid.UUID) ([]*domain.Meta, error) {
	slog.Debug("Get Meta Data by user", slog.String("user ID", userID.String()))

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
	if err != nil && !errors.Is(err, sql.ErrNoRows) && rows.Err() != nil {
		return nil, fmt.Errorf("could not execute get meta query: %w", err)
	}
	defer rows.Close()

	var res []*domain.Meta
	for rows.Next() {
		row := &domain.Meta{}
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

// DeleteMetaDataByID removes a metadata record from the metas table by its unique ID. Returns an error if the operation fails.
func (s *Storage) DeleteMetaDataByID(id uuid.UUID) error {
	slog.Debug("Delete Meta Data by ID", slog.String("ID", id.String()))

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

// Close terminates the database connection and releases any associated resources. Returns an error if it fails.
func (s *Storage) Close() error {
	return s.db.Close()
}

// Ping checks the connection to the database and returns an error if the database is not reachable.
func (s *Storage) Ping() error {
	return s.db.Ping()
}
