package storage

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage/psql"
)

type Commands interface {
	SaveUser(*models.UserData) error
	SaveText(string) error
	GetText(uuid.UUID) (*models.TextData, error)
	SaveBankCard(*models.BankCardData) error
	GetBankCard(uuid.UUID) (*models.BankCardData, error)
	SaveMetaData(*models.Meta) error
	GetMetaData(uuid.UUID) (*models.Meta, error)
	Close() error
}

func NewInstance(cfg *config.DB) (Commands, error) {
	conn, err := psql.New(cfg.Address, cfg.Name, cfg.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres database instance: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database instance: %w", err)
	}

	return conn, nil
}
