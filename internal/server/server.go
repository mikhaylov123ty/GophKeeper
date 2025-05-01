package server

import (
	"fmt"
	"net"

	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc/handlers"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage"
)

// Server - структура сервера
type Server struct {
	grpc *grpc.GRPCServer
	auth *auth
}

type auth struct {
	cryptoKey string
	hashKey   string
}

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

func (s *Server) Start() error {
	listen, err := net.Listen("tcp", config.GetAddress().String())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return s.grpc.Server.Serve(listen)
}
