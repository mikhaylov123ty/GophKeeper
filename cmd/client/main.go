package main

import (
	"fmt"

	clientConfig "github.com/mikhaylov123ty/GophKeeper/internal/client/config"
)

func main() {
	config, err := clientConfig.New()
	if err != nil {
		panic(err)
	}
	fmt.Printf("config initialized %+v\n", config)

	fmt.Println("Hello i'm client")
}
