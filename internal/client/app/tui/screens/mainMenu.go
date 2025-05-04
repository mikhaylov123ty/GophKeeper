package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

const (
	ViewOption = "View items"
	AddOption  = "Add an item"
	BackOption = "Back"
)

// MainMenu represents the main interface of the UI, allowing navigation through categories and managing transitions.
// categories holds the list of available menu categories for user selection.
// cursor tracks the current position in the category list for navigation purposes.
// itemsManager is responsible for handling item and metadata interactions within the application.
// nextScreen defines the next screen or view to transition into based on the user's actions.
type MainMenu struct {
	categories   []string
	cursor       int
	itemsManager models.ItemsManager
	nextScreen   models.Screen
}

// NewMainMenu initializes a MainMenu instance with a list of categories and an ItemsManager for handling items.
func NewMainMenu(categories []string, itemsManager models.ItemsManager) *MainMenu {
	return &MainMenu{
		categories:   categories,
		itemsManager: itemsManager,
	}
}

// Update handles user input to navigate the main menu, modify the cursor, or transition to the next screen or quit the application.
func (m MainMenu) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.cursor = (m.cursor + 1) % len(m.categories)
		case "up":
			m.cursor = (m.cursor - 1 + len(m.categories)) % len(m.categories)
		case "enter":
			category := m.categories[m.cursor]
			if category == ExitCategory {
				return m, tea.Quit // Exit the application
			}
			m.nextScreen = &ActionsMenu{
				options:      []string{ViewOption, AddOption, BackOption},
				category:     category,
				itemsManager: m.itemsManager,
				backScreen:   m,
			}
			return m.nextScreen, nil
		}
	}

	return m, nil
}

// View renders the main menu as a string, highlighting the selected category and appending navigation instructions.
func (m MainMenu) View() string {
	s := utils.TitleStyle.Render("Main Menu:\n\n")
	for i, category := range m.categories {
		if m.cursor == i {
			s += utils.CursorStyle.Render(fmt.Sprintf("[x] %s\n", category)) // Selected option with color
		} else {
			s += utils.UnselectedStyle.Render(fmt.Sprintf("[ ] %s\n", category)) // Unselected option with color
		}
	}
	s += utils.NavigateFooter()
	return s
}
