package server

import (
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc"
	"github.com/sirupsen/logrus"

	"net"
)

// Server - структура сервера
type Server struct {
	services services
	logger   *logrus.Logger
	auth     *auth
}

// services - структура команд БД и файла с бэкапом
type services struct {
	gRPCStorageCommands *grpc.StorageCommands
}

type auth struct {
	cryptoKey string
	hashKey   string
}

func New(gRPCStorageCommands *grpc.StorageCommands) *Server {
	return &Server{
		services: services{
			gRPCStorageCommands: gRPCStorageCommands,
		},
	}
}

func (s *Server) Start() error {
	listen, err := net.Listen("tcp", config.GetAddress().GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	//grpc server
	gRPC := grpc.NewServer(config.GetKeys().HashKey, config.GetKeys().CryptoKey, s.services.gRPCStorageCommands)

	return gRPC.Server.Serve(listen)
}
