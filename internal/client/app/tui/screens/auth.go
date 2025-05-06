package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

// AuthScreen represents a screen for handling user authentication in a terminal-based UI application.
// It manages the input fields for username and password, navigation, and authentication logic.
type AuthScreen struct {
	username     string
	password     string
	next         models.Screen
	itemsManager models.ItemsManager
	cursor       int
}

// NewAuthScreen creates an AuthScreen instance wrapped in a Model, initializing it with the next screen and items manager.
func NewAuthScreen(next models.Screen, itemsManager models.ItemsManager) *models.Model {
	return &models.Model{
		CurrentScreen: &AuthScreen{
			next:         next,
			itemsManager: itemsManager,
		},
	}
}

// Update processes incoming messages, updates the AuthScreen's state, and returns the next screen and an optional command.
func (s *AuthScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			switch s.cursor {
			case 0: // username
				if len(s.username) > 0 {
					s.username = s.username[:len(s.username)-1]
				}
			case 1: // password
				if len(s.password) > 0 {
					s.password = s.password[:len(s.password)-1]
				}
			}

		case tea.KeyTab:
			s.cursor = (s.cursor + 1) % 2

		case tea.KeyEnter:
			if err := s.itemsManager.PostUserData(s.username, s.password); err != nil {
				return &ErrorScreen{
					backScreen: s,
					err:        err,
				}, nil
			}

			if err := s.itemsManager.SyncMeta(); err != nil {
				return &ErrorScreen{
					backScreen: s,
					err:        err,
				}, nil
			}

			return s.next, nil
		case tea.KeyCtrlQ:
			return s, tea.Quit

		default:
			if len(msg.String()) == 1 && msg.String() != "\x00" {
				// Handle character input based on the currently focused field
				if s.cursor == 0 { // Username field
					s.username += msg.String()
				} else if s.cursor == 1 { // Password field
					s.password += msg.String()
				}
			}
		}
	}

	return s, nil
}

// View generates a string representation of the AuthScreen for rendering, including input fields and footer instructions.
func (s *AuthScreen) View() string {
	var sb strings.Builder

	sb.WriteString(utils.TitleStyle.Render("Please log in:\n"))

	// Render Username Field
	sb.WriteString(fmt.Sprintf("\nUsername: %s\n", utils.SelectedStyle.Render(s.username)))
	// Render Password Field (masked)
	sb.WriteString(fmt.Sprintf("Password: %s\n", utils.SelectedStyle.Render(strings.Repeat("•", len(s.password)))))
	// Render Footer
	sb.WriteString(utils.AuthFooter())

	return sb.String()
}

// Init method for AuthScreen
func (s *AuthScreen) Init() tea.Cmd {
	return nil // No command to run initially
}
