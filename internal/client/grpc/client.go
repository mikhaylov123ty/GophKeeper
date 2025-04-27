package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type Client struct {
	JWTToken string
	Handlers *Handlers
}

type Handlers struct {
	ItemDataHandler pb.ItemDataHandlersClient
	MetaDataHandler pb.MetaDataHandlersClient
	AuthHandler     pb.UserHandlersClient
}

func New() (*Client, error) {
	var err error
	instance := Client{}

	//TODO add interceptors
	interceptors := []grpc.UnaryClientInterceptor{
		instance.withJWT,
	}

	tlsCred, err := credentials.NewClientTLSFromFile(config.GetKeys().PublicCert, "localhost")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	conn, err := grpc.NewClient(
		config.GetAddress().String(),
		grpc.WithTransportCredentials(tlsCred),
		grpc.WithChainUnaryInterceptor(interceptors...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating grpc client: %w", err)
	}

	instance.Handlers = &Handlers{
		ItemDataHandler: pb.NewItemDataHandlersClient(conn),
		MetaDataHandler: pb.NewMetaDataHandlersClient(conn),
		AuthHandler:     pb.NewUserHandlersClient(conn),
	}

	slog.Info("Client Created", slog.String("address", config.GetAddress().String()))

	return &instance, nil
}

func (c *Client) withJWT(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	if c.JWTToken != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("Authorization", c.JWTToken))
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
