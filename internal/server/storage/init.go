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
	SaveText(*models.TextData) error
	GetTextByID(uuid.UUID) (*models.TextData, error)
	SaveBankCard(*models.BankCardData) error
	GetBankCardById(uuid.UUID) (*models.BankCardData, error)
	SaveMetaData(*models.Meta) error
	GetMetaDataByUser(uuid.UUID) ([]*models.Meta, error)
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
