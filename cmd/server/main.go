package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/mikhaylov123ty/GophKeeper/internal/server"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage"
	"github.com/mikhaylov123ty/GophKeeper/pkg/logger"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	log.Printf("config initialized %+v", *cfg)

	if err := logger.Init(
		config.GetLogger().LogLevel,
		config.GetLogger().LogFormat,
	); err != nil {
		panic(err)
	}

	slog.Info("logger initialized",
		slog.String("level", config.GetLogger().LogLevel),
		slog.String("format", config.GetLogger().LogFormat),
	)

	storageService, err := storage.NewInstance(config.GetDB())
	if err != nil {
		panic(err)
	}
	defer storageService.Close()

	slog.Info("storage initialized",
		slog.String("driver", "postgres"),
		slog.String("dsn", config.GetDB().Address),
		slog.String("db_name", config.GetDB().Name),
	)

	serverInstance := server.New(storageService)

	go func() {
		if err = serverInstance.Start(); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Hello i'm server")
}
