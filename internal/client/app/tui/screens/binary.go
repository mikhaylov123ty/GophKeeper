package screens

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	dbModels "github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

const (
	binaryFields = 3
)

type viewBinaryDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.BinaryData
}

type addBinaryItemScreen struct {
	*models.ItemScreen
	selectedItem *models.MetaItem
	newItemData  *dbModels.Binary
}

func (screen *viewBinaryDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			ext, err := mime.ExtensionsByType(screen.itemData.ContentType)
			if err != nil {
				return &ErrorScreen{
					backScreen: screen,
					err:        err,
				}, nil
			}

			outputFile, err := os.Create(strings.Join([]string{config.GetOutputFolder(), screen.itemData.Name, ext[0]}, ""))
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

		case "q":
			return screen.backScreen, nil
		}
	}
	return screen, nil
}

func (screen *viewBinaryDataScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sFile Information%s\n"+
			"=======================%s"+
			"%sTitle: %s%s\n"+
			"%sType: %s%s\n",
		utils.ColorBold, utils.ColorReset,
		separator,
		utils.ColorGreen, screen.itemData.Name, utils.ColorReset,
		utils.ColorYellow, screen.itemData.ContentType, utils.ColorReset,
	) + utils.BinaryItemDataFooter()
}

func (screen *addBinaryItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	var err error
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.NewTitle != "" && screen.NewDesc != "" && screen.newItemData.FilePath != "" {
				filename := filepath.Base(screen.newItemData.FilePath)
				extension := filepath.Ext(screen.newItemData.FilePath)
				name := filename[:len(filename)-len(extension)]

				screen.newItemData.Binary.Content, err = os.ReadFile(screen.newItemData.FilePath)
				if err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}
				screen.newItemData.Binary.ContentType = mime.TypeByExtension(extension)
				screen.newItemData.Binary.Name = name

				if screen.newItemData != nil {
					binaryData, err := json.Marshal(screen.newItemData.Binary)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					var id uuid.UUID
					var dataID string
					if screen.selectedItem != nil {
						id = screen.selectedItem.Id
						dataID = screen.selectedItem.DataID
					} else {
						id = uuid.New()
					}

					newItem := models.MetaItem{
						Id:          id,
						Title:       screen.NewTitle,
						Description: screen.NewDesc,
					}

					metaData := pb.MetaData{
						Id:          newItem.Id.String(),
						Title:       newItem.Title,
						Description: newItem.Description,
						DataType:    screen.Category,
					}

					resp, err := screen.ItemsManager.PostItemData(binaryData, dataID, &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					if screen.selectedItem != nil {
						screen.selectedItem.Title = screen.NewTitle
						screen.selectedItem.Description = screen.NewDesc
						screen.selectedItem.Modified = resp.Modified
					} else {
						newItem.DataID = resp.DataId
						newItem.Created = resp.Created
						newItem.Modified = resp.Modified

						screen.ItemsManager.SaveMetaItem(screen.Category, &newItem)
					}
				}
			}
			return screen.BackScreen, nil // Go back to category menu

		case "ctrl+q": // Go back to the previous menu
			return screen.BackScreen, nil
		case "up":
			screen.Cursor = (screen.Cursor - 1 + binaryFields) % binaryFields // Focus on Title
		case "down":
			screen.Cursor = (screen.Cursor + 1) % binaryFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

func (screen *addBinaryItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &dbModels.Binary{}
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
	styles[screen.Cursor] = utils.SelectedStyle
	// Build each line
	addLine("Title:", screen.NewTitle, styles[0])
	addLine("Description:", screen.NewDesc, styles[1])
	addLine("File Path:", screen.newItemData.FilePath, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addBinaryItemScreen) handleInput(input string) {
	fields := []string{screen.NewTitle, screen.NewDesc, screen.newItemData.FilePath}

	// Backspace logic
	if input == "backspace" {
		if len(fields[screen.Cursor]) > 0 {
			fields[screen.Cursor] = fields[screen.Cursor][:len(fields[screen.Cursor])-1]
		}
	} else {
		// Ignore special keys
		if input != "up" && input != "down" && input != "esc" {
			fields[screen.Cursor] += input
		}
	}

	// Update the fields back to the screen state
	screen.NewTitle = fields[0]
	screen.NewDesc = fields[1]
	screen.newItemData.FilePath = fields[2]
}
