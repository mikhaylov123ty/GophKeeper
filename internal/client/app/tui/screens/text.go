package screens

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	dbModels "github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

const (
	textFields = 3
)

type viewTextDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.TextData
}

type addTextItemScreen struct {
	*models.ItemScreen
	selectedItem *models.MetaItem
	newItemData  *dbModels.TextData
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
	) + utils.ItemDataFooter()
}

func (screen *addTextItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.NewTitle != "" && screen.NewDesc != "" && screen.newItemData.Text != "" {
				// Create new item and add to the manager
				if screen.newItemData != nil {
					textData, err := json.Marshal(screen.newItemData)
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

					resp, err := screen.ItemsManager.PostItemData(textData, dataID, &metaData)
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

			return screen.BackScreen, nil

		case "ctrl+q":
			return screen.BackScreen, nil

		case "up":
			screen.Cursor = (screen.Cursor - 1 + textFields) % textFields

		case "down":
			screen.Cursor = (screen.Cursor + 1) % textFields

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

	// Set styles based on Cursor position
	styles := []lipgloss.Style{
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
	}
	styles[screen.Cursor] = utils.SelectedStyle

	// Build each line
	addLine("Title:", screen.NewTitle, styles[0])
	addLine("Description:", screen.NewDesc, styles[1])
	addLine("Text:", screen.newItemData.Text, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addTextItemScreen) handleInput(input string) {
	fields := []string{screen.NewTitle, screen.NewDesc, screen.newItemData.Text}

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
	screen.newItemData.Text = fields[2]
}
