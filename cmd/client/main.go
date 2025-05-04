// The provided `main.go` file represents the entry point of the `GophKeeper` client's application.

package main

import (
	"log/slog"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app"
	grpcClient "github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"

	clientConfig "github.com/mikhaylov123ty/GophKeeper/internal/client/config"
)

// The main function serves as the starting point of the application execution
func main() {
	config, err := clientConfig.New()
	if err != nil {
		panic(err)
	}

	slog.Info("config initialized",
		slog.String("Address", config.Address.String()),
		slog.String("Config File", config.ConfigFile),
		slog.String("Cert", config.Keys.PublicCert),
		slog.String("Output Folder", config.OutputFolder),
	)

	grpc, err := grpcClient.New()
	if err != nil {
		panic(err)
	}

	appSvc := app.New(grpc)

	if err = appSvc.Run(); err != nil {
		panic(err)
	}
}
