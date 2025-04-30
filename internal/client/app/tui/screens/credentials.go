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
	credsFields = 4
)

type viewCredsDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.CredsData
}

type addCredsItemScreen struct {
	*models.ItemScreen
	newItemData *dbModels.CredsData
}

type editCredsItemScreen struct {
	*models.ItemScreen
	selectedItem *models.MetaItem
	newItemData  *dbModels.CredsData
}

func (screen *viewCredsDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return screen.backScreen, nil
		}
	}

	return screen, nil
}

// TODO UNIFY
func (screen *viewCredsDataScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sCreds Information%s\n"+
			"=======================%s"+
			"Login: %s%s\n"+
			"Password: %s%s\n",
		utils.ColorBold, utils.ColorReset,
		separator,
		utils.ColorGreen, screen.itemData.Login,
		utils.ColorGreen, screen.itemData.Password,
	)
}

func (screen *addCredsItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.NewTitle != "" && screen.NewDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          uuid.New(),
					Title:       screen.NewTitle,
					Description: screen.NewDesc,
				}

				if screen.newItemData != nil {
					credsData, err := json.Marshal(screen.newItemData)
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
						DataType:    screen.Category,
					}

					resp, err := screen.ItemsManager.PostItemData(credsData, "", &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					newItem.DataID = resp.DataId
					newItem.Created = resp.Created
					newItem.Modified = resp.Modified

					screen.ItemsManager.SaveMetaItem(screen.Category, &newItem)
				}
			}
			return screen.BackScreen, nil // Go back to category menu

		case "ctrl+q": // Go back to the previous menu
			return screen.BackScreen, nil
		case "up":
			screen.Cursor = (screen.Cursor - 1 + credsFields) % credsFields // Focus on Title
		case "down":
			screen.Cursor = (screen.Cursor + 1) % credsFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

// TODO UNIFY THIS
func (screen *addCredsItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &dbModels.CredsData{}
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
	styles[screen.Cursor] = utils.SelectedStyle

	// Build each line
	addLine("Title:", screen.NewTitle, styles[0])
	addLine("Description:", screen.NewDesc, styles[1])
	addLine("Login:", screen.newItemData.Login, styles[2])
	addLine("Password:", screen.newItemData.Password, styles[3])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addCredsItemScreen) handleInput(input string) {
	fields := []string{screen.NewTitle, screen.NewDesc, screen.newItemData.Login, screen.newItemData.Password}

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
	screen.newItemData.Login = fields[2]
	screen.newItemData.Password = fields[3]
}

// TODO DEPRECATE? THISI SAME AS ADD BUT FILLED ITEMS
func (screen *editCredsItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.NewTitle != "" && screen.NewDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          screen.selectedItem.Id,
					Title:       screen.NewTitle,
					Description: screen.NewDesc,
				}

				if screen.newItemData != nil {
					//TODO create dedicated func
					credsData, err := json.Marshal(screen.newItemData)
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
						DataType:    screen.Category,
					}

					resp, err := screen.ItemsManager.PostItemData(credsData, screen.selectedItem.DataID, &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					screen.selectedItem.Title = screen.NewTitle
					screen.selectedItem.Description = screen.NewDesc
					screen.selectedItem.Modified = resp.Modified
				}

				return screen.BackScreen, nil // Go back to category menu
			}
		case "ctrl+q": // Go back to the previous menu
			return screen.BackScreen, nil
		case "up":
			screen.Cursor = (screen.Cursor - 1 + credsFields) % credsFields // Focus on Title
		case "down":
			screen.Cursor = (screen.Cursor + 1) % credsFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

func (screen *editCredsItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &dbModels.CredsData{}
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
	styles[screen.Cursor] = utils.SelectedStyle

	// Build each line
	addLine("Title:", screen.NewTitle, styles[0])
	addLine("Description:", screen.NewDesc, styles[1])
	addLine("Login:", screen.newItemData.Login, styles[2])
	addLine("Password:", screen.newItemData.Password, styles[3])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	// Add instructions at the end
	instructions := "Press Enter to save, Q to cancel, or Backspace to delete the last character."

	return fmt.Sprintf("%s\n\n%s\n", result, instructions)
}

func (screen *editCredsItemScreen) handleInput(input string) {
	fields := []string{screen.NewTitle, screen.NewDesc, screen.newItemData.Login, screen.newItemData.Password}

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
	screen.newItemData.Login = fields[2]
	screen.newItemData.Password = fields[3]
}
