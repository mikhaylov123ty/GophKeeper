package screens

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	dbModels "github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

const (
	binaryFields = 3
)

type viewBinaryDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.BinaryData
}

type binaryItemScreen struct {
	itemsManager models.ItemsManager
	category     string
	newTitle     string
	newDesc      string
	newItemData  *dbModels.Binary
	createdTime  string
	cursor       int
	backScreen   models.Screen
}

type addBinaryItemScreen struct {
	*binaryItemScreen
}

// TODO remove unused fields
type editBinaryItemScreen struct {
	*binaryItemScreen
	selectedItem *models.MetaItem
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
	)
}

func (screen *addBinaryItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	var err error
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.FilePath != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          uuid.New(),
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}

				filename := filepath.Base(screen.newItemData.FilePath)
				extension := filepath.Ext(screen.newItemData.FilePath)
				name := filename[:len(filename)-len(extension)]

				screen.newItemData.Binary.Content, err = os.ReadFile(screen.newItemData.FilePath)
				screen.newItemData.Binary.ContentType = mime.TypeByExtension(extension)
				screen.newItemData.Binary.Name = name

				if err != nil {
					//TODO error screen
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}
				if screen.newItemData != nil {
					binaryData, err := json.Marshal(screen.newItemData.Binary)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					metaData := pb.MetaData{
						Id:          newItem.Id.String(),
						Title:       newItem.Title,
						Description: newItem.Description,
						DataType:    screen.category,
					}

					resp, err := screen.itemsManager.PostItemData(binaryData, "", &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					newItem.DataID = resp.DataId
					newItem.Created = resp.Created
					newItem.Modified = resp.Modified

					//TODO store to local storage
					screen.itemsManager.SaveMetaItem(screen.category, &newItem)
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
	styles[screen.cursor] = utils.SelectedStyle
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

func (screen *editBinaryItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	var err error
	if keyMsg, ok := msg.(tea.KeyMsg); ok {

		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.FilePath != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          screen.selectedItem.Id,
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}

				filename := filepath.Base(screen.newItemData.FilePath)
				extension := filepath.Ext(screen.newItemData.FilePath)
				name := filename[:len(filename)-len(extension)]

				screen.newItemData.Binary.Content, err = os.ReadFile(screen.newItemData.FilePath)
				screen.newItemData.Binary.ContentType = mime.TypeByExtension(extension)
				screen.newItemData.Binary.Name = name

				if err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}

				if screen.newItemData != nil {
					//TODO create dedicated func
					binaryData, err := json.Marshal(screen.newItemData.Binary)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					metaData := pb.MetaData{
						Id:          screen.selectedItem.Id.String(),
						Title:       newItem.Title,
						Description: newItem.Description,
						DataType:    screen.category,
					}

					resp, err := screen.itemsManager.PostItemData(binaryData, screen.selectedItem.DataID, &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					screen.selectedItem.Title = screen.newTitle
					screen.selectedItem.Description = screen.newDesc
					screen.selectedItem.Modified = resp.Modified
				}

				return screen.backScreen, nil // Go back to category menu
			}
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

func (screen *editBinaryItemScreen) View() string {
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
	styles[screen.cursor] = utils.SelectedStyle
	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("File Path:", screen.newItemData.FilePath, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *editBinaryItemScreen) handleInput(input string) {
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
