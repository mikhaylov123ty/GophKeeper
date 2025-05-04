package screens

import (
	"encoding/json"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

const (
	binaryFields = 3
	toMB         = 1048576
)

type viewBinaryDataScreen struct {
	backScreen models.Screen
	itemData   *models.BinaryData
}

type addBinaryItemScreen struct {
	*itemScreen
	newItemData *models.Binary
}

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

func (screen *addBinaryItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	var err error
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.FilePath != "" {
				filename := filepath.Base(screen.newItemData.FilePath)
				extension := filepath.Ext(screen.newItemData.FilePath)
				name := filename[:len(filename)-len(extension)]

				screen.newItemData.Binary.Content, err = os.ReadFile(screen.newItemData.FilePath)
				if err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        fmt.Errorf("filepath: %s, error: %w", screen.newItemData.FilePath, err),
					}, nil
				}

				screen.newItemData.Binary.Name = strings.Join([]string{name, extension}, "")
				screen.newItemData.Binary.FileSize = float64(len(string(screen.newItemData.Binary.Content))) / toMB

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

func (screen *addBinaryItemScreen) handleInput(input string) {
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
}
