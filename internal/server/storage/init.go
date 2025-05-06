package storage

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mikhaylov123ty/GophKeeper/internal/domain"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage/psql"
)

// Commands defines database operations for managing users, items, and metadata, including CRUD and lifecycle methods.
type Commands interface {
	SaveUser(*domain.UserData) error
	GetUserByLogin(string) (*domain.UserData, error)
	SaveItemData(*domain.ItemData, *domain.Meta) error
	GetItemDataByID(uuid.UUID) (*domain.ItemData, error)
	DeleteItemDataByID(uuid.UUID) error
	GetMetaDataByUser(uuid.UUID) ([]*domain.Meta, error)
	DeleteMetaDataByID(uuid.UUID) error
	Close() error
}

// NewInstance initializes a new database instance with the provided configuration and applies migrations.
func NewInstance(cfg *config.DB) (Commands, error) {
	slog.Debug("db config", slog.Any("cfg", *cfg))
	conn, err := psql.New(cfg.DSN, cfg.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres database instance: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database instance: %w", err)
	}

	return conn, nil
}
