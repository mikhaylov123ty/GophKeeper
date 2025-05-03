package main

import (
	"log"
	"log/slog"
	"sync"

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
		slog.String("dsn", config.GetDB().DSN),
	)

	serverInstance, err := server.New(storageService)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting server", slog.String("address", config.GetAddress().String()))

		if err = serverInstance.Start(); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
