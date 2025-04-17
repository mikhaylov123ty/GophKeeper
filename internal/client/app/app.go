package app

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
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
	tui := tui.NewItemManager(a.grpcClient)
	prog := tea.NewProgram(tui)
	if _, err := prog.Run(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
	//for {
	//	scanner := bufio.NewScanner(os.Stdin)
	//	scanner.Scan()
	//	err := scanner.Err()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	fmt.Println("INPUT:", scanner.Text())
	//
	//	switch scanner.Text() {
	//	case "add text":
	//		fmt.Println("Enter Text")
	//		scanner.Scan()
	//		err = scanner.Err()
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//
	//		a.grpcClient.PostText(ctx, scanner.Text())
	//	}
	//}
	return nil
}
