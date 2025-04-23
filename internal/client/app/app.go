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
	itemManager, err := tui.NewItemManager(a.grpcClient)
	if err != nil {
		return fmt.Errorf("could not create tui: %w", err)
	}
	prog := tea.NewProgram(itemManager)
	if _, err = prog.Run(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
	return nil
}
