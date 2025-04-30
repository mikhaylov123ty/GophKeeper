package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

type AuthScreen struct {
	username    string
	password    string
	next        models.Screen
	itemManager models.ItemsManager
	cursor      int
}

// NewAuthScreen initializes the AuthScreen
func NewAuthScreen(next models.Screen, itemManager models.ItemsManager) *models.Model {
	return &models.Model{
		CurrentScreen: &AuthScreen{
			next:        next,
			itemManager: itemManager,
		},
	}
}

// Update method for AuthScreen
func (s *AuthScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
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

		case tea.KeyTab:
			s.cursor = (s.cursor + 1) % 2

		case tea.KeyEnter:
			if err := s.itemManager.PostUserData(s.username, s.password); err != nil {
				return &ErrorScreen{
					backScreen: s,
					err:        err,
				}, nil
			}

			if err := s.itemManager.SyncMeta(); err != nil {
				return &ErrorScreen{
					backScreen: s,
					err:        err,
				}, nil
			}

			return s.next, nil
		case tea.KeyCtrlQ:
			return s, tea.Quit

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
	sb.WriteString(utils.AuthFooter())

	return sb.String()
}

// Init method for AuthScreen
func (s *AuthScreen) Init() tea.Cmd {
	return nil // No command to run initially
}
