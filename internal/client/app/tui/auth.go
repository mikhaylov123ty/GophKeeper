package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type AuthScreen struct {
	username    string
	password    string
	next        Screen
	itemManager *ItemManager
	cursor      int
}

// NewAuthScreen initializes the AuthScreen
func NewAuthScreen(next Screen, itemManager *ItemManager) *Model {
	return &Model{
		currentScreen: &AuthScreen{
			next:        next,
			itemManager: itemManager,
		},
	}
}

// Update method for AuthScreen
func (s *AuthScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			switch s.cursor {
			case 0:
				if len(s.username) > 0 {
					s.username = s.username[:len(s.username)-1]
				}
			case 1:
				if len(s.password) > 0 {
					s.password = s.password[:len(s.password)-1]
				}
			}
			return s, nil
		case tea.KeyTab:
			s.cursor = (s.cursor + 1) % 2
			return s, nil
		case tea.KeyCtrlQ:
			return s, tea.Quit // Exit on ESC
		}
		switch msg.String() {
		//case "tab": // Change focus between fields

		case "enter": // Submit the form
			if err := s.login(); err != nil {
				return &ErrorScreen{
					backScreen: s,
					err:        err,
				}, nil
			}

			if err := s.itemManager.syncMeta(); err != nil {
				return &ErrorScreen{
					backScreen: s,
					err:        err,
				}, nil
			}

			return s.next, nil // Go to main menu if authenticated

		default:
			// Handle character input based on the currently focused field
			if s.cursor == 0 { // Username field
				s.username += msg.String()
			} else if s.cursor == 1 { // Password field
				s.password += msg.String()
			}
		}
	}

	return s, nil
}

// View method for AuthScreen
func (s *AuthScreen) View() string {
	var sb strings.Builder

	// Basic styles
	usernameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	passwordStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	sb.WriteString("Please log in:\n\n")

	// Render Username Field
	sb.WriteString(fmt.Sprintf("Username: %s\n", usernameStyle.Render(s.username)))
	// Render Password Field (masked)
	sb.WriteString(fmt.Sprintf("Password: %s\n", passwordStyle.Render(strings.Repeat("•", len(s.password)))))

	// Render Footer
	sb.WriteString(authFooter())

	return sb.String()
}

// Init method for AuthScreen
func (s *AuthScreen) Init() tea.Cmd {
	return nil // No command to run initially
}

func (s *AuthScreen) login() error {
	res, err := s.itemManager.grpcClient.Handlers.AuthHandler.PostUserData(context.Background(), &pb.PostUserDataRequest{
		Login:    s.username,
		Password: s.password,
	})
	if err != nil {
		return fmt.Errorf("failed login: %w", err)
	}

	if res.Error != "" {
		return fmt.Errorf("failed login: %s", res.Error)
	}

	if res.UserId == "" {
		return fmt.Errorf("failed login: empty user id")
	}

	s.itemManager.userID = res.UserId
	s.itemManager.grpcClient.JWTToken = res.Jwt

	return nil
}

func authFooter() string {
	return backgroundStyle.Render(separatorStyle.Render("\nPress Tab to switch fields, Enter to submit, or CTRL+Q to exit.\n"))
}
