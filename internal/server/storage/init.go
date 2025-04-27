package storage

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage/psql"
)

type Commands interface {
	SaveUser(*models.UserData) error
	GetUserByLogin(string) (*models.UserData, error)
	SaveItemData(data *models.ItemData) error
	GetItemDataByID(uuid.UUID) (*models.ItemData, error)
	DeleteItemDataByID(uuid.UUID) error
	SaveMetaData(*models.Meta) error
	GetMetaDataByUser(uuid.UUID) ([]*models.Meta, error)
	DeleteMetaDataById(uuid.UUID) error
	Close() error
}

func NewInstance(cfg *config.DB) (Commands, error) {
	slog.Debug("db config", slog.Any("cfg", *cfg))
	conn, err := psql.New(cfg.DSN, cfg.Name, cfg.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres database instance: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database instance: %w", err)
	}

	return conn, nil
}
