package app

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
)

type App struct {
	grpcClient *grpc.Client
}

// New - Creates and initializes a new instance of the `App` struct, which serves as the central service of the application
func New(grpcClient *grpc.Client) *App {
	return &App{
		grpcClient: grpcClient,
	}
}

// Run - Executes the main flow of the application,
// initializing and running the terminal-based user interface (TUI)
func (a *App) Run() error {
	itemManager, err := tui.NewItemManager(a.grpcClient)
	if err != nil {
		return fmt.Errorf("could not create tui: %w", err)
	}

	prog := tea.NewProgram(itemManager)
	if _, err = prog.Run(); err != nil {
		return fmt.Errorf("could not create tea program: %w", err)
	}

	return nil
}
