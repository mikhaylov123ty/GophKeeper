// The provided `main.go` file represents the entry point of the `GophKeeper` server's application.

package main

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/mikhaylov123ty/GophKeeper/internal/server"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage"
	"github.com/mikhaylov123ty/GophKeeper/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

// The main function serves as the starting point of the application execution
func main() {
	fmt.Printf("Server Build Version: %s\n", buildVersion)
	fmt.Printf("Server Build Date: %s\n", buildDate)

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

	//TODO add graceful shutdown
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err = serverInstance.Start(); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
