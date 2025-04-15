package main

import (
	"context"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"

	clientConfig "github.com/mikhaylov123ty/GophKeeper/internal/client/config"
)

func main() {
	config, err := clientConfig.New()
	if err != nil {
		panic(err)
	}
	fmt.Printf("config initialized %+v\n", config)

	fmt.Println("Hello i'm client")

	client, err := grpc.NewClient()
	if err != nil {
		panic(err)
	}
	appSvc := app.New(client)

	if err = appSvc.Run(context.Background()); err != nil {
		panic(err)
	}
}
