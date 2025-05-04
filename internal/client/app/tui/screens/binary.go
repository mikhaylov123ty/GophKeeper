package screens

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
)

const (
	binaryFields = 3
	mB           = 1048576
	contentLimit = 30
)

// viewBinaryDataScreen represents a screen for viewing binary data content in a terminal-based UI application.
// backScreen stores the previous screen to return to after exiting the current view.
// itemData holds the binary data and its associated metadata for display or operations.
type viewBinaryDataScreen struct {
	backScreen models.Screen
	itemData   *models.BinaryData
}

// addBinaryItemScreen represents a screen used for adding or editing binary items, such as files, with metadata details.
type addBinaryItemScreen struct {
	*itemScreen
	newItemData *models.Binary
}

// Update processes a message, handling key events and managing transitions between screens or error handling.
func (screen *viewBinaryDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			outputFile, err := os.Create(strings.Join([]string{config.GetOutputFolder(), screen.itemData.Name}, ""))
			if err != nil {
				return &ErrorScreen{
					backScreen: screen,
					err:        err,
				}, nil
			}
			defer outputFile.Close()

			if _, err = outputFile.Write(screen.itemData.Content); err != nil {
				return &ErrorScreen{
					backScreen: screen,
					err:        err,
				}, nil
			}

			return screen.backScreen, nil

		case "ctrl+q":
			return screen.backScreen, nil
		}
	}
	return screen, nil
}

// View returns a string representation of the binary data screen, including its metadata and instructions for interaction.
func (screen *viewBinaryDataScreen) View() string {
	body := utils.DataHeader()

	body += fmt.Sprintf(
		"\n%sTitle: %s%s\n"+
			"%sSize: %.2f MB%s\n",
		utils.ColorGreen, screen.itemData.Name, utils.ColorReset,
		utils.ColorRed, screen.itemData.FileSize, utils.ColorReset,
	)

	body += utils.BinaryItemDataFooter()

	return body
}

// Update processes user input messages, handles navigation, data validation, and posting for the addBinaryItemScreen.
func (screen *addBinaryItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	var err error
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.FilePath != "" {
				filePath := filepath.Clean(screen.newItemData.FilePath)
				filename := filepath.Base(filePath)
				extension := filepath.Ext(filePath)
				name := filename[:len(filename)-len(extension)]

				screen.newItemData.Binary.Content, err = os.ReadFile(filePath)
				if err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        fmt.Errorf("filepath: %s, error: %w", filePath, err),
					}, nil
				}

				screen.newItemData.Binary.Name = strings.Join([]string{name, extension}, "")
				screen.newItemData.Binary.FileSize = float64(len(string(screen.newItemData.Binary.Content))) / mB

				if screen.newItemData.Binary.FileSize > float64(contentLimit) {
					return &ErrorScreen{
						backScreen: screen,
						err:        fmt.Errorf("file size is too big: %.2fMB, max size is %dMB", screen.newItemData.Binary.FileSize, contentLimit),
					}, nil
				}

				if screen.newItemData != nil {
					var binaryData []byte
					binaryData, err = json.Marshal(screen.newItemData.Binary)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					if err = screen.postItemData(binaryData); err != nil {
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
			screen.cursor = (screen.cursor - 1 + binaryFields) % binaryFields // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % binaryFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

// View generates and returns the styled string representation of the addBinaryItemScreen for rendering.
func (screen *addBinaryItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.Binary{}
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
	}
	styles[screen.cursor] = utils.CursorStyle
	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("File Path:", screen.newItemData.FilePath, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

// handleInput processes user keyboard input and updates the state of the screen, including text fields and navigation.
func (screen *addBinaryItemScreen) handleInput(input string) {
	if input == "\x00" {
		return
	}

	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.FilePath}

	// Backspace logic
	if input == "backspace" {
		if len(fields[screen.cursor]) > 0 {
			fields[screen.cursor] = fields[screen.cursor][:len(fields[screen.cursor])-1]
		}
	} else {
		// Ignore special keys
		if input != "up" && input != "down" && input != "esc" {
			fields[screen.cursor] += input
		}
	}

	// Update the fields back to the screen state
	screen.newTitle = fields[0]
	screen.newDesc = fields[1]
	screen.newItemData.FilePath = fields[2]

	if screen.cursor == 2 && screen.newItemData.FilePath != "" {
		if screen.newItemData.FilePath[0] == '[' && screen.newItemData.FilePath[len(screen.newItemData.FilePath)-1] == ']' {
			screen.newItemData.FilePath = screen.newItemData.FilePath[1 : len(screen.newItemData.FilePath)-1]
		}
	}
}
