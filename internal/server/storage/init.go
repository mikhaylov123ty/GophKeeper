package storage

import (
	"fmt"

	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage/psql"
)

type Storage struct {
	GRPCStorageCommands *grpc.StorageCommands
	Close               func() error
}

func NewInstance(cfg *config.DB) (*Storage, error) {
	conn, err := psql.New(cfg.Address, cfg.Name, cfg.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres database instance: %w", err)
	}

	return &Storage{
			GRPCStorageCommands: grpc.NewStorageCommands(conn),
			Close:               conn.Close,
		},
		nil
}
