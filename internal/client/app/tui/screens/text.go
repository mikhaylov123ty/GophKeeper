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
	textFields = 3
)

type viewTextDataScreen struct {
	backScreen models.Screen
	itemData   *models.TextData
}

type addTextItemScreen struct {
	*itemScreen
	newItemData *models.TextData
}

func (screen *viewTextDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			return screen.backScreen, nil
		}
	}
	return screen, nil
}

func (screen *viewTextDataScreen) View() string {
	body := utils.DataHeader()

	body += fmt.Sprintf(
		"\n%sText: %s%s\n",
		utils.ColorGreen, screen.itemData.Text, utils.ColorReset,
	)

	body += utils.ItemDataFooter()

	return body
}

func (screen *addTextItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Text != "" {
				// Create new item and add to the manager
				if screen.newItemData != nil {
					textData, err := json.Marshal(screen.newItemData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					if err = screen.postItemData(textData); err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}
				}
			}

			return screen.backScreen, nil

		case "ctrl+q":
			return screen.backScreen, nil

		case "up":
			screen.cursor = (screen.cursor - 1 + textFields) % textFields

		case "down":
			screen.cursor = (screen.cursor + 1) % textFields

		default:
			screen.handleInput(keyMsg.String())
		}
	}

	return screen, nil
}

func (screen *addTextItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.TextData{}
	}

	// Define an array of elements to hold the rendered strings
	var lines []string

	// Define a function for creating the styled label lines
	addLine := func(label string, value string, style lipgloss.Style) {
		lines = append(lines, fmt.Sprintf("%s %s", style.Render(label), style.Render(value)))
	}

	// Set styles based on Cursor position
	styles := []lipgloss.Style{
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
	}
	styles[screen.cursor] = utils.CursorStyle

	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("Text:", screen.newItemData.Text, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addTextItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.Text}

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
	screen.newItemData.Text = fields[2]
}
