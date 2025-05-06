package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

// Client represents a gRPC client with JWT authorization and handlers for various services.
type Client struct {
	JWTToken string
	Handlers *Handlers
}

// Handlers is a struct containing clients for interacting with various gRPC services.
// ItemDataHandler interacts with services handling item data operations.
// MetaDataHandler interacts with services handling metadata operations.
// AuthHandler interacts with services handling user authentication operations.
type Handlers struct {
	ItemDataHandler pb.ItemDataHandlersClient
	MetaDataHandler pb.MetaDataHandlersClient
	AuthHandler     pb.UserHandlersClient
}

// New initializes and returns a new gRPC client with TLS credentials and middleware for JWT authentication.
func New() (*Client, error) {
	var err error
	instance := Client{}

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

	slog.Info("Client Created", slog.String("address", ":"+config.GetAddress().GRPCPort))

	return &instance, nil
}

// withJWT adds a JWT token to the gRPC request context as an Authorization header if the token is set in the client.
// It then invokes the given gRPC method using the provided invoker and options.
func (c *Client) withJWT(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	if c.JWTToken != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("Authorization", c.JWTToken))
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
