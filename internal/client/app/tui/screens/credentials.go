package screens

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

const (
	credsFields = 4
)

type viewCredsDataScreen struct {
	backScreen models.Screen
	itemData   *models.CredsData
}

type addCredsItemScreen struct {
	*itemScreen
	newItemData *models.CredsData
}

func (screen *viewCredsDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			return screen.backScreen, nil
		}
	}

	return screen, nil
}

func (screen *viewCredsDataScreen) View() string {
	body := utils.DataHeader()

	body += fmt.Sprintf(
		"\n%sLogin: %s%s\n"+
			"%sPassword: %s%s\n",
		utils.ColorGreen, screen.itemData.Login, utils.ColorReset,
		utils.ColorGreen, screen.itemData.Password, utils.ColorReset,
	)
	body += utils.ItemDataFooter()

	return body
}

func (screen *addCredsItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				if screen.newItemData != nil {
					credsData, err := json.Marshal(screen.newItemData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					if err = screen.postItemData(credsData); err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}
				}
			}
			return screen.backScreen, nil // Go back to category menu

		case "ctrl+q": // Go back to the previous menu
			return screen.backScreen, nil
		case "up":
			screen.cursor = (screen.cursor - 1 + credsFields) % credsFields // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % credsFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

func (screen *addCredsItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.CredsData{}
	}

	// Define an array of elements to hold the rendered strings
	var lines []string

	// Define a function for creating the styled label lines
	addLine := func(label string, value string, style lipgloss.Style) {
		lines = append(lines, fmt.Sprintf("%s %s", style.Render(label), style.Render(value)))
	}

	// Set styles based on cursor position
	styles := []lipgloss.Style{
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
	}
	styles[screen.cursor] = utils.CursorStyle

	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("Login:", screen.newItemData.Login, styles[2])
	addLine("Password:", screen.newItemData.Password, styles[3])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addCredsItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.Login, screen.newItemData.Password}

	// Backspace logic
	if input == "backspace" {
		if len(fields[screen.cursor]) > 0 {
			fields[screen.cursor] = fields[screen.cursor][:len(fields[screen.cursor])-1]
		}
	} else {
		// Ignore special keys
		if len(input) == 1 {
			fields[screen.cursor] += input
		}
	}

	// Update the fields back to the screen state
	screen.newTitle = fields[0]
	screen.newDesc = fields[1]
	screen.newItemData.Login = fields[2]
	screen.newItemData.Password = fields[3]
}
