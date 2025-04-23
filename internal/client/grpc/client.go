package grpc

import (
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	TextHandler      pb.TextHandlersClient
	MetaHandler      pb.MetaDataHandlersClient
	BankCardsHandler pb.BankCardHandlersClient
}

func NewClient() (*Client, error) {
	conn, err := grpc.NewClient(
		config.GetAddress().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating grpc client: %w", err)
	}

	fmt.Println("ADDRESS", config.GetAddress().String())

	return &Client{
		TextHandler:      pb.NewTextHandlersClient(conn),
		MetaHandler:      pb.NewMetaDataHandlersClient(conn),
		BankCardsHandler: pb.NewBankCardHandlersClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.Close()
}
