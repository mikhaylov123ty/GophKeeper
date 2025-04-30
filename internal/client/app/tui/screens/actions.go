package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

const (
	TextCategory  = "Text"
	CredsCategory = "Creds"
	FileCategory  = "Files"
	CardCategory  = "Cards"
	ExitCategory  = "Exit" // New exit category
)

type ActionsMenu struct {
	options      []string
	cursor       int
	category     string
	itemsManager models.ItemsManager
	backScreen   models.Screen
}

func (cm ActionsMenu) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			cm.cursor = (cm.cursor + 1) % len(cm.options)
		case "up":
			cm.cursor = (cm.cursor - 1 + len(cm.options)) % len(cm.options)
		case "enter":
			switch cm.cursor {
			case 0: // View items
				return &viewMetaItemsScreen{category: cm.category, itemsManager: cm.itemsManager, backScreen: cm}, nil
			case 1: // Add item
				switch cm.category {
				case TextCategory:
					return &addTextItemScreen{
						ItemScreen: &models.ItemScreen{
							ItemsManager: cm.itemsManager,
							BackScreen:   cm,
							Category:     cm.category,
						},
					}, nil

				case CredsCategory:
					return &addCredsItemScreen{
						ItemScreen: &models.ItemScreen{
							ItemsManager: cm.itemsManager,
							BackScreen:   cm,
							Category:     cm.category,
						},
					}, nil

				case FileCategory:
					return &addBinaryItemScreen{
						ItemScreen: &models.ItemScreen{
							ItemsManager: cm.itemsManager,
							BackScreen:   cm,
							Category:     cm.category,
						},
					}, nil

				case CardCategory:
					return &addBankCardItemScreen{
						ItemScreen: &models.ItemScreen{
							ItemsManager: cm.itemsManager,
							BackScreen:   cm,
							Category:     cm.category,
						},
					}, nil
				}
			case 2: // Back
				return cm.backScreen, nil
			}
		case "esc": // Go back to main menu
			return cm.backScreen, nil
		}
	}
	return cm, nil
}

func (cm ActionsMenu) View() string {
	s := utils.TitleStyle.Render("Category Menu:\n\n")
	for i, option := range cm.options {
		if cm.cursor == i {
			s += utils.SelectedStyle.Render(fmt.Sprintf("[x] %s\n", option)) // Selected option with color
		} else {
			s += utils.UnselectedStyle.Render(fmt.Sprintf("[ ] %s\n", option)) // Unselected option with color
		}
	}

	s += utils.NavigateFooter()
	return s
}
