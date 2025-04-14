package grpc

import (
	"context"
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
	auth   *auth
	Server *grpc.Server
}

type auth struct {
	cryptoKey string
	hashKey   string
}

// NewServer создает инстанс gRPC сервера
func NewServer(cryptoKey string, hashKey string,
	textHandler *handlers.TextHandler,
	bankCardHandler *handlers.BankCardDataHandler,
	metaDataHandler *handlers.MetaDataHandler,

) *GRPCServer {
	instance := &GRPCServer{
		auth: &auth{
			cryptoKey: cryptoKey,
			hashKey:   hashKey,
		},
	}

	// Определение перехватчиков
	interceptors := []grpc.UnaryServerInterceptor{
		instance.withLogger,
		//instance.withHash,
		instance.withAuth,
	}

	//Регистрация инстанса gRPC с перехватчиками
	instance.Server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...))

	pb.RegisterTextHandlersServer(instance.Server, textHandler)
	pb.RegisterBankCardHandlersServer(instance.Server, bankCardHandler)
	pb.RegisterMetaDataHandlersServer(instance.Server, metaDataHandler)

	return instance
}

// withLogger - перехватчик логирует запросы
func (g *GRPCServer) withLogger(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// Запуск таймера
	start := time.Now()

	slog.InfoContext(ctx, "gRPC server received request", slog.String("method", info.FullMethod))

	// Запуск RPC-метода
	resp, err = handler(ctx, req)

	// Логирует код и таймер
	e, _ := status.FromError(err)
	slog.InfoContext(ctx, "Request completed ", slog.String("code", e.Code().String()), slog.Any("time spent", time.Since(start)))

	return resp, err
}

func (g *GRPCServer) withAuth(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "can't extract metadata from request")
	}

	header, ok := meta["JWT"]
	if ok {
		return nil, status.Error(codes.Unauthenticated, "can't found JWT header")
	}

	token, err := jwt.Parse(header[0], func(token *jwt.Token) (interface{}, error) {
		return config.GetKeys().CryptoKey, nil
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "can't parse JWT header")
	}

	if !token.Valid {
		return nil, status.Error(codes.PermissionDenied, "JWT token is invalid")
	}

	return handler(ctx, req)
}

//// withHash - перехватчик проверяет наличие хеша в метаданных и сверяет с телом запроса
//func (g *GRPCServer) withHash(ctx context.Context, req any,
//	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
//	// Проверка наличия флага ключа
//	if g.auth.hashKey != "" {
//		g.logger.Infof("start checking gRPC request hash")
//
//		// Чтеные метаданных
//		meta, ok := metadata.FromIncomingContext(ctx)
//		if !ok {
//			return nil, status.Errorf(codes.Internal, "can't extract metadata from request")
//		}
//		var requestHeader []byte
//		header, ok := meta["hashsha256"]
//		if !ok {
//			return nil, status.Errorf(codes.Unauthenticated, "can't extract hash header from request")
//		}
//		requestHeader, err = hex.DecodeString(header[0])
//		if err != nil {
//			return nil, status.Errorf(codes.InvalidArgument, "can't decode hash header from request")
//		}
//
//		// Чтение тела запроса
//		body := req.(*pb.PostUpdatesRequest).Metrics
//
//		// Вычисление и валидация хеша
//		hash := utils.GetHash(g.auth.hashKey, body)
//		if !hmac.Equal(hash, requestHeader) {
//			return nil, status.Errorf(codes.PermissionDenied, "hash does not match")
//		}
//	}
//
//	return handler(ctx, req)
//}
