package app

import (
	"bufio"
	"context"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
	"log"
	"os"
)

type App struct {
	grpcClient *grpc.Client
}

func New(grpcClient *grpc.Client) *App {
	return &App{
		grpcClient: grpcClient,
	}
}

func (a *App) Run(ctx context.Context) error {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("INPUT:", scanner.Text())

		switch scanner.Text() {
		case "add text":
			fmt.Println("Enter Text")
			scanner.Scan()
			err = scanner.Err()
			if err != nil {
				log.Fatal(err)
			}

			a.grpcClient.PostText(ctx, scanner.Text())
		}
	}
}
