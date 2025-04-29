package screens

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	dbModels "github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"strings"
)

type viewTextDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.TextData
}
type textItemScreen struct {
	itemsManager models.ItemsManager
	category     string
	newTitle     string
	newDesc      string
	newItemData  *dbModels.TextData
	createdTime  string
	cursor       int
	backScreen   models.Screen
}
type addTextItemScreen struct {
	*textItemScreen
}

// TODO remove unused fields
type editTextItemScreen struct {
	*textItemScreen
	selectedItem *models.MetaItem
}

func (screen *viewTextDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return screen.backScreen, nil
		}
	}
	return screen, nil
}

// TODO move to models and unify
func (screen *viewTextDataScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sText Information%s\n"+
			"=======================%s"+
			"%sText: %s%s\n",
		utils.ColorBold, utils.ColorReset,
		separator,
		utils.ColorGreen, screen.itemData.Text, utils.ColorReset,
	)
}

func (screen *addTextItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Text != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          uuid.New(),
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}

				if screen.newItemData != nil {
					textData, err := json.Marshal(screen.newItemData)
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

					resp, err := screen.itemsManager.PostItemData(textData, "", &metaData)
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
			screen.cursor = (screen.cursor - 1 + 3) % 3 // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % 3 // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

// TODO UNIFY THIS, like run build func in method
func (screen *addTextItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &dbModels.TextData{}
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
		if input != "up" && input != "down" && input != "esc" {
			fields[screen.cursor] += input
		}
	}

	// Update the fields back to the screen state
	screen.newTitle = fields[0]
	screen.newDesc = fields[1]
	screen.newItemData.Text = fields[2]
}

func (screen *editTextItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Text != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          screen.selectedItem.Id,
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}

				if screen.newItemData != nil {
					//TODO create dedicated func
					textData, err := json.Marshal(screen.newItemData)
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

					resp, err := screen.itemsManager.PostItemData(textData, screen.selectedItem.DataID, &metaData)
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
			screen.cursor = (screen.cursor - 1 + 3) % 3 // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % 3 // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

func (screen *editTextItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &dbModels.TextData{}
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
	addLine("Text:", screen.newItemData.Text, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *editTextItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.Text}

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
	screen.newItemData.Text = fields[2]
}
