package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc/handlers"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage"
)

// Server represents a gRPC server with authentication capabilities, managing GRPCServer and auth configurations.
type Server struct {
	grpc *grpc.GRPCServer
	auth *auth
}

// auth represents authentication configuration, managing cryptographic and hashing keys for secure operations.
type auth struct {
	cryptoKey string
	hashKey   string
}

// New initializes and returns a new Server instance configured with the provided storage commands, or an error if setup fails.
func New(storageCommands storage.Commands) (*Server, error) {
	gRPC, err := grpc.NewServer(
		handlers.NewTextHandler(storageCommands, storageCommands),
		handlers.NewMetaDataHandler(storageCommands, storageCommands),
		handlers.NewAuthHandler(storageCommands, storageCommands),
	)
	if err != nil {
		return nil, fmt.Errorf("failed build new server: %w", err)
	}

	return &Server{
		grpc: gRPC,
	}, nil
}

// Start initializes the server listener and starts serving gRPC requests on the configured network address.
func (s *Server) Start() error {
	slog.Info("starting server", slog.String("address", config.GetAddress().String()))

	listen, err := net.Listen("tcp", config.GetAddress().String())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return s.grpc.Server.Serve(listen)
}
