package grpc

import (
	"context"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Client struct {
	TextHandler pb.TextHandlersClient
}

func NewClient() (*Client, error) {
	conn, err := grpc.NewClient(
		config.GetAddress().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating grpc client: %w", err)
	}

	return &Client{TextHandler: pb.NewTextHandlersClient(conn)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) PostText(ctx context.Context, text string) error {
	resp, err := c.TextHandler.PostTextData(ctx, &pb.PostTextDataRequest{Text: text})
	if err == nil {
		return nil
	}
	if e, ok := status.FromError(err); ok {
		switch e.Code() {
		case codes.Unavailable:
			return fmt.Errorf("server unavailable: %w", err)
		default:
			return fmt.Errorf("post updates: Code: %s, Message: %s", e.Code(), e.Message())
		}
	} else {
		log.Printf("Can't parse error: %s\n", err.Error())
	}
}
