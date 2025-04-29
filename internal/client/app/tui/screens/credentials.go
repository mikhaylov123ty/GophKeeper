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

const (
	credsFields = 4
)

type viewCredsDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.CredsData
}

type credsItemScreen struct {
	itemsManager models.ItemsManager
	category     string
	newTitle     string
	newDesc      string
	newItemData  *dbModels.CredsData
	cursor       int
	backScreen   models.Screen
}

type addCredsItemScreen struct {
	*credsItemScreen
}

type editCredsItemScreen struct {
	*credsItemScreen
	selectedItem *models.MetaItem
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
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          uuid.New(),
					Title:       screen.newTitle,
					Description: screen.newDesc,
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
						DataType:    screen.category,
					}

					resp, err := screen.itemsManager.PostItemData(credsData, "", &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					newItem.DataID = resp.DataId
					newItem.Created = resp.Created
					newItem.Modified = resp.Modified

					screen.itemsManager.SaveMetaItem(screen.category, &newItem)
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
	styles[screen.cursor] = utils.SelectedStyle

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
		if input != "up" && input != "down" && input != "esc" {
			fields[screen.cursor] += input
		}
	}

	// Update the fields back to the screen state
	screen.newTitle = fields[0]
	screen.newDesc = fields[1]
	screen.newItemData.Login = fields[2]
	screen.newItemData.Password = fields[3]
}

// TODO DEPRECATE? THISI SAME AS ADD BUT FILLED ITEMS
func (screen *editCredsItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				// Create new item and add to the manager
				newItem := models.MetaItem{
					Id:          screen.selectedItem.Id,
					Title:       screen.newTitle,
					Description: screen.newDesc,
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
						DataType:    screen.category,
					}

					resp, err := screen.itemsManager.PostItemData(credsData, screen.selectedItem.DataID, &metaData)
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
			screen.cursor = (screen.cursor - 1 + credsFields) % credsFields // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % credsFields // Focus on Description
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
	styles[screen.cursor] = utils.SelectedStyle

	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("Login:", screen.newItemData.Login, styles[2])
	addLine("Password:", screen.newItemData.Password, styles[3])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	// Add instructions at the end
	instructions := "Press Enter to save, Q to cancel, or Backspace to delete the last character."

	return fmt.Sprintf("%s\n\n%s\n", result, instructions)
}

func (screen *editCredsItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.Login, screen.newItemData.Password}

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
	screen.newItemData.Login = fields[2]
	screen.newItemData.Password = fields[3]
}
