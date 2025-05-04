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

// viewTextDataScreen represents a screen displaying a specific textual data item.
// backScreen stores the previous screen for navigation purposes.
// itemData holds the textual data to be displayed.
type viewTextDataScreen struct {
	backScreen models.Screen
	itemData   *models.TextData
}

// addTextItemScreen represents a screen for adding or editing text-based items.
// It extends itemScreen and includes specific data handling for TextData items.
type addTextItemScreen struct {
	*itemScreen
	newItemData *models.TextData
}

// Update processes incoming messages, handles key events, and returns the updated screen and an optional command.
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

// View renders the content of the viewTextDataScreen, displaying the data header, the text field, and a styled footer.
func (screen *viewTextDataScreen) View() string {
	body := utils.DataHeader()

	body += fmt.Sprintf(
		"\n%sText: %s%s\n",
		utils.ColorGreen, screen.itemData.Text, utils.ColorReset,
	)

	body += utils.ItemDataFooter()

	return body
}

// Update processes incoming messages or events, updates the screen's state, and returns the updated screen and command.
func (screen *addTextItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
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

					// Post item data to server
					if err = screen.postItemData(textData); err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}
				}
			}

			return screen.backScreen, nil

		case tea.KeyCtrlQ:
			return screen.backScreen, nil

		case tea.KeyUp:
			screen.cursor = (screen.cursor - 1 + textFields) % textFields

		case tea.KeyDown:
			screen.cursor = (screen.cursor + 1) % textFields

		default:
			screen.handleInput(keyMsg.String())
		}
	}

	return screen, nil
}

// View renders the UI for the addTextItemScreen, displaying input fields and highlighting the current cursor position.
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

// handleInput processes user input to update the current field based on the cursor position and modifies screen state.
func (screen *addTextItemScreen) handleInput(input string) {
	if input == "\x00" {
		return
	}

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
