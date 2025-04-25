package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type Client struct {
	//	Conn     *grpc.ClientConn
	JWTToken string
	Handlers *Handlers
}

type Handlers struct {
	TextHandler      pb.TextHandlersClient
	MetaHandler      pb.MetaDataHandlersClient
	BankCardsHandler pb.BankCardHandlersClient
	AuthHandler      pb.UserHandlersClient
}

func New() (*Client, error) {
	var err error
	instance := Client{}

	//TODO add interceptors
	interceptors := []grpc.UnaryClientInterceptor{
		instance.withJWT,
	}

	conn, err := grpc.NewClient(
		config.GetAddress().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptors...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating grpc client: %w", err)
	}

	instance.Handlers = &Handlers{
		TextHandler:      pb.NewTextHandlersClient(conn),
		MetaHandler:      pb.NewMetaDataHandlersClient(conn),
		BankCardsHandler: pb.NewBankCardHandlersClient(conn),
		AuthHandler:      pb.NewUserHandlersClient(conn),
	}

	fmt.Println("ADDRESS", config.GetAddress().String())

	return &instance, nil
}

func (c *Client) withJWT(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	if c.JWTToken != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("Authorization", c.JWTToken))
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
