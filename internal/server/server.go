package server

import (
	"fmt"

	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc/handlers"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/storage"

	"net"
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

func New(storageCommands storage.Commands) *Server {
	gRPC := grpc.NewServer(
		config.GetKeys().HashKey,
		config.GetKeys().CryptoKey,
		handlers.NewTextHandler(storageCommands, storageCommands),
		handlers.NewBankCardDataHandler(storageCommands, storageCommands),
		handlers.NewMetaDataHandler(storageCommands, storageCommands),
	)

	return &Server{
		grpc: gRPC,
	}
}

func (s *Server) Start() error {
	listen, err := net.Listen("tcp", config.GetAddress().GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return s.grpc.Server.Serve(listen)
}
