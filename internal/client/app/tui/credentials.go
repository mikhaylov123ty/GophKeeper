package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

const (
	credsFields = 4
)

type ViewCredsItemsScreen struct {
	options     []string
	category    Category
	itemManager *ItemManager
	backScreen  Screen
	list        *list.Model
}

type ViewCredsDataScreen struct {
	backScreen Screen
	itemData   *models.CredsData
}

// TODO optimize duplications
type AddCredsItemScreen struct {
	itemManager *ItemManager
	category    Category
	newTitle    string
	newDesc     string
	newItemData *models.CredsData
	createdTime string
	cursor      int
	backScreen  Screen
}

// TODO remove unused fields
type EditCredsItemScreen struct {
	itemManager  *ItemManager
	category     Category
	newTitle     string
	newDesc      string
	selectedItem *MetaItem
	newItemData  *models.CredsData
	createdTime  string
	cursor       int
	backScreen   Screen
}

func (screen *ViewCredsItemsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			screen.list.CursorDown()
		case "up":
			screen.list.CursorUp()
		case "enter":
			if len(screen.itemManager.metaItems[screen.category]) > 0 {
				itemDataID := screen.list.SelectedItem().(*MetaItem).dataID
				itemData, err := screen.itemManager.getItemData(itemDataID)
				if err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}

				var credsData models.CredsData
				if err = json.Unmarshal([]byte(itemData), &credsData); err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}

				return &ViewCredsDataScreen{
					backScreen: screen,
					itemData: &models.CredsData{
						Login:    credsData.Login,
						Password: credsData.Password,
					},
				}, nil
			}
		case "d":
			if len(screen.itemManager.metaItems[screen.category]) > 0 {
				//todo failed screen
				if err := screen.itemManager.deleteItem(
					screen.list.SelectedItem().(*MetaItem).Id,
					screen.category,
					screen.list.SelectedItem().(*MetaItem).dataID,
				); err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}
				screen.list.CursorUp()
			}
		case "e":
			if screen.itemManager.metaItems[screen.category] != nil {
				return &EditCredsItemScreen{itemManager: screen.itemManager,
					backScreen:   screen,
					category:     screen.category,
					selectedItem: screen.list.SelectedItem().(*MetaItem),
					newTitle:     screen.list.SelectedItem().(*MetaItem).Title,
					newDesc:      screen.list.SelectedItem().(*MetaItem).Description,
				}, nil
			}
		case "q":
			return screen.backScreen, nil // Go back when ESC is pressed
		}
	}
	return screen, nil
}

func (screen *ViewCredsItemsScreen) View() string {
	if screen.list == nil {
		listModel := list.New([]list.Item{}, metaItemDelegate{}, 10, listHeight)
		screen.list = &listModel
	}

	if len(screen.itemManager.metaItems[screen.category]) == 0 {
		return cursorStyle.Render("No items to display.\n\nPress ESC to go back.\n")
	}

	listItems := []list.Item{}
	for _, v := range screen.itemManager.metaItems[screen.category] {
		listItems = append(listItems, v)
	}

	screen.list.SetItems(listItems)

	return screen.list.View()
}

func (screen *ViewCredsDataScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return screen.backScreen, nil
		}
	}
	return screen, nil
}

func (screen *ViewCredsDataScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sCreds Information%s\n"+
			"=======================%s"+
			"Login: %s%s\n"+
			"Password: %s%s\n",
		ColorBold, ColorReset,
		separator,
		ColorGreen, screen.itemData.Login,
		ColorGreen, screen.itemData.Password,
	)
}

func (screen *AddCredsItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				// Create new item and add to the manager
				newItem := MetaItem{
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
						DataType:    string(screen.category),
						UserId:      screen.itemManager.userID,
					}

					resp, err := screen.itemManager.postItemData(credsData, "", &metaData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					newItem.dataID = resp.DataId
					newItem.Created = resp.Created
					newItem.Modified = resp.Modified

					//TODO store to local storage
					screen.itemManager.metaItems[screen.category] = append(screen.itemManager.metaItems[screen.category], &newItem)
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

func (screen *AddCredsItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.CredsData{}
	}

	// Define an array of elements to hold the rendered strings
	var lines []string

	// Define a function for creating the styled label lines
	addLine := func(label string, value string, style lipgloss.Style) {
		lines = append(lines, fmt.Sprintf("%s %s", style.Render(label), style.Render(value)))
	}

	// Set styles based on cursor position
	styles := []lipgloss.Style{unselectedStyle, unselectedStyle, unselectedStyle, unselectedStyle}
	styles[screen.cursor] = selectedStyle // Highlight the currently focused element

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

func (screen *AddCredsItemScreen) handleInput(input string) {
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

func (screen *EditCredsItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" && screen.newItemData.Login != "" && screen.newItemData.Password != "" {
				// Create new item and add to the manager
				newItem := MetaItem{
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
						DataType:    string(screen.category),
						UserId:      screen.itemManager.userID,
					}

					resp, err := screen.itemManager.postItemData(credsData, screen.selectedItem.dataID, &metaData)
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

func (screen *EditCredsItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.CredsData{}
	}

	// Define an array of elements to hold the rendered strings
	var lines []string

	// Define a function for creating the styled label lines
	addLine := func(label string, value string, style lipgloss.Style) {
		lines = append(lines, fmt.Sprintf("%s %s", style.Render(label), style.Render(value)))
	}

	// Set styles based on cursor position
	styles := []lipgloss.Style{unselectedStyle, unselectedStyle, unselectedStyle, unselectedStyle}
	styles[screen.cursor] = selectedStyle // Highlight the currently focused element

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

func (screen *EditCredsItemScreen) handleInput(input string) {
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
