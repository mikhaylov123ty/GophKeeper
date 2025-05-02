package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/grpc/handlers"
)

// GRPCServer - структура инстанса gRPC сервера
type GRPCServer struct {
	Server *grpc.Server
}

// NewServer создает инстанс gRPC сервера
func NewServer(
	itemsDataHandler *handlers.ItemsDataHandler,
	metaDataHandler *handlers.MetaDataHandler,
	authHandler *handlers.AuthHandler,
) (*GRPCServer, error) {
	instance := &GRPCServer{}

	// Определение перехватчиков
	interceptors := []grpc.UnaryServerInterceptor{
		instance.withLogger,
		instance.withAuth,
	}

	creds, err := credentials.NewServerTLSFromFile("public.crt", "private.key")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	//Регистрация инстанса gRPC с перехватчиками
	instance.Server = grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(interceptors...),
	)

	pb.RegisterItemDataHandlersServer(instance.Server, itemsDataHandler)
	pb.RegisterMetaDataHandlersServer(instance.Server, metaDataHandler)
	pb.RegisterUserHandlersServer(instance.Server, authHandler)

	return instance, nil
}

// withLogger - перехватчик логирует запросы
func (g *GRPCServer) withLogger(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Запуск таймера
	start := time.Now()

	slog.InfoContext(ctx, "gRPC server received request", slog.String("method", info.FullMethod))
	slog.InfoContext(ctx, "gRPC server received request", slog.Any("req", req))

	// Запуск RPC-метода
	resp, err = handler(ctx, req)

	// Логирует код и таймер
	e, _ := status.FromError(err)
	slog.InfoContext(ctx, "Request completed ", slog.String("code", e.Code().String()), slog.Any("time spent", time.Since(start)))

	return resp, err
}

func (g *GRPCServer) withAuth(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if config.GetKeys().JWTKey != "" {
		if info.FullMethod != "/server_grpc.UserHandlers/PostUserData" && config.GetKeys().JWTKey != "" {
			slog.InfoContext(ctx, "starting verifying JWT")
			meta, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				slog.ErrorContext(ctx, "Failed to get metadata")
				return nil, status.Error(codes.Internal, "can't extract metadata from request")
			}

			header, ok := meta["authorization"]
			if !ok {
				slog.ErrorContext(ctx, "Failed to get Authorization header")
				return nil, status.Error(codes.Unauthenticated, "can't found JWT header")
			}

			token, err := jwt.Parse(header[0], func(token *jwt.Token) (interface{}, error) {
				return []byte(config.GetKeys().JWTKey), nil
			})
			if err != nil {
				slog.ErrorContext(ctx, "Failed to parse Authorization header")
				return nil, status.Error(codes.Unauthenticated, "can't parse Authorization header")
			}

			if !token.Valid {
				slog.ErrorContext(ctx, "JWT token is invalid")
				return nil, status.Error(codes.PermissionDenied, "JWT token is invalid")
			}
		}
	}

	return handler(ctx, req)
}
