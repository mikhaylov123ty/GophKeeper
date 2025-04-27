package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

const (
	cardFields = 5
)

type ViewBankCardItemsScreen struct {
	options      []string
	category     Category
	itemManager  *ItemManager
	editScreen   Screen
	deleteScreen Screen
	backScreen   Screen
	list         *list.Model
}

type ViewBankCardDataScreen struct {
	backScreen Screen
	itemData   *models.BankCardData
}

type AddBankCardItemScreen struct {
	itemManager *ItemManager
	category    Category
	newTitle    string
	newDesc     string
	newItemData *models.BankCardData
	cursor      int
	backScreen  Screen
}

type EditBankCardItemScreen struct {
	itemManager  *ItemManager
	category     Category
	newTitle     string
	newDesc      string
	selectedItem *MetaItem
	newItemData  *models.BankCardData
	cursor       int
	backScreen   Screen
}

func (screen *ViewBankCardItemsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			screen.list.CursorDown()
		case "up":
			screen.list.CursorUp()
		case "enter":
			itemDataID := screen.list.SelectedItem().(*MetaItem).dataID
			itemData, err := screen.itemManager.getItemData(itemDataID)
			if err != nil {
				return ViewBankCardDataScreen{
					backScreen: screen,
					itemData: &models.BankCardData{
						CardNum: err.Error(),
					},
				}, nil
			}
			var cardData models.BankCardData
			if err := json.Unmarshal([]byte(itemData), &cardData); err != nil {
				return ViewBankCardDataScreen{
					backScreen: screen,
					itemData: &models.BankCardData{
						CardNum: err.Error(),
					},
				}, nil
			}
			return ViewBankCardDataScreen{
				backScreen: screen,
				itemData: &models.BankCardData{
					CardNum: cardData.CardNum,
					Expiry:  cardData.Expiry,
					CVV:     cardData.CVV,
				},
			}, nil
		case "e":
			if screen.itemManager.metaItems[screen.category] != nil {
				return &EditBankCardItemScreen{itemManager: screen.itemManager,
					backScreen:   screen,
					category:     screen.category,
					selectedItem: screen.list.SelectedItem().(*MetaItem),
					newTitle:     screen.list.SelectedItem().(*MetaItem).Title,
					newDesc:      screen.list.SelectedItem().(*MetaItem).Description,
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
					return screen, nil
				}
				screen.list.CursorUp()
			}
		case "q":
			return screen.backScreen, nil // Go back when ESC is pressed
		}
	}
	return screen, nil
}

func (screen *ViewBankCardItemsScreen) View() string {
	if screen.list == nil {
		listModel := list.New([]list.Item{}, metaItemDelegate{}, 10, listHeight)
		screen.list = &listModel
	}

	if len(screen.itemManager.metaItems[screen.category]) == 0 {
		return "No items to display.\n\nPress Q to go back.\n"
	}

	listItems := []list.Item{}
	for _, v := range screen.itemManager.metaItems[screen.category] {
		listItems = append(listItems, v)
	}

	screen.list.SetItems(listItems)
	screen.list.SetShowHelp(false)
	screen.list.Title = "Bank Cards List"

	s := screen.list.View()
	s += backgroundStyle.Render(separatorStyle.Render("\nUse arrow keys to navigate. e to edit. d to delete. enter to select.\n")) // Navigation instructions

	return s
}

func (screen ViewBankCardDataScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return screen.backScreen, nil
		}
	}

	return screen, nil
}

func (screen ViewBankCardDataScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sCard Information%s\n"+
			"=======================%s"+
			"%sCard Num: %s%s\n"+
			"%sExpiry: %s%s\n"+
			"%sCVV: %s%s\n",
		ColorBold, ColorReset, separator,
		ColorGreen, screen.itemData.CardNum, ColorReset,
		ColorYellow, screen.itemData.Expiry, ColorReset,
		ColorRed, screen.itemData.CVV, ColorReset,
	)
}

func (screen *AddBankCardItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" {
				// Create new item and add to the manager
				newItem := MetaItem{
					Id:          uuid.New(),
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}
				if screen.newItemData != nil {
					cardData, err := json.Marshal(screen.newItemData)
					if err != nil {
						return screen, nil
					}

					metaData := pb.MetaData{
						Id:          newItem.Id.String(),
						Title:       newItem.Title,
						Description: newItem.Description,
						DataType:    string(screen.category),
						UserId:      screen.itemManager.userID,
					}

					resp, err := screen.itemManager.postItemData(cardData, "", &metaData)
					if err != nil {
						return screen.backScreen, func() tea.Msg {
							newItem.Title = err.Error()
							screen.itemManager.metaItems[screen.category] = append(screen.itemManager.metaItems[screen.category], &newItem)

							return fmt.Sprintf("ERROR %s,", err.Error())
						}
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
			screen.cursor = (screen.cursor - 1 + cardFields) % cardFields // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % cardFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}
	}
	return screen, nil
}

func (screen *AddBankCardItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.BankCardData{}
	}

	// Define an array of elements to hold the rendered strings
	var lines []string

	// Define a function for creating the styled label lines
	addLine := func(label string, value string, style lipgloss.Style) {
		lines = append(lines, fmt.Sprintf("%s %s", style.Render(label), style.Render(value)))
	}

	// Set styles based on cursor position
	styles := []lipgloss.Style{unselectedStyle, unselectedStyle, unselectedStyle, unselectedStyle, unselectedStyle}
	styles[screen.cursor] = selectedStyle // Highlight the currently focused element

	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("Card Num:", screen.newItemData.CardNum, styles[2])
	addLine("Expiry:", screen.newItemData.Expiry, styles[3])
	addLine("CVV:", screen.newItemData.CVV, styles[4])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	// Add instructions at the end
	instructions := "Press Enter to save, Q to cancel, or Backspace to delete the last character."

	return fmt.Sprintf("%s\n\n%s\n", result, instructions)
}

func (screen *AddBankCardItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.CardNum, screen.newItemData.Expiry, screen.newItemData.CVV}

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
	screen.newItemData.CardNum = fields[2]
	screen.newItemData.Expiry = fields[3]
	screen.newItemData.CVV = fields[4]
}

func (screen *EditBankCardItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" &&
				screen.newDesc != "" &&
				screen.newItemData.CardNum != "" &&
				screen.newItemData.Expiry != "" &&
				screen.newItemData.CVV != "" {

				// Create new item and add to the manager
				newItem := MetaItem{
					Id:          screen.selectedItem.Id,
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}

				if screen.newItemData != nil {
					//TODO create dedicated func
					cardData, err := json.Marshal(screen.newItemData)
					if err != nil {
						return screen, nil
					}

					metaData := pb.MetaData{
						Id:          newItem.Id.String(),
						Title:       newItem.Title,
						Description: newItem.Description,
						DataType:    string(screen.category),
						UserId:      screen.itemManager.userID,
					}

					resp, err := screen.itemManager.postItemData(cardData, screen.selectedItem.dataID, &metaData)
					if err != nil {
						return screen.backScreen, func() tea.Msg {
							newItem.Title = err.Error()
							screen.itemManager.metaItems[screen.category] = append(screen.itemManager.metaItems[screen.category], &newItem)

							return fmt.Sprintf("ERROR %s,", err.Error())
						}
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
			screen.cursor = (screen.cursor - 1 + cardFields) % cardFields // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % cardFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

func (screen *EditBankCardItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &models.BankCardData{}
	}

	// Define an array of elements to hold the rendered strings
	var lines []string

	// Define a function for creating the styled label lines
	addLine := func(label string, value string, style lipgloss.Style) {
		lines = append(lines, fmt.Sprintf("%s %s", style.Render(label), style.Render(value)))
	}

	// Set styles based on cursor position
	styles := []lipgloss.Style{unselectedStyle, unselectedStyle, unselectedStyle, unselectedStyle, unselectedStyle}
	styles[screen.cursor] = selectedStyle // Highlight the currently focused element

	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("Card Num:", screen.newItemData.CardNum, styles[2])
	addLine("Expiry:", screen.newItemData.Expiry, styles[3])
	addLine("CVV:", screen.newItemData.CVV, styles[4])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	// Add instructions at the end
	instructions := "Press Enter to save, Q to cancel, or Backspace to delete the last character."

	return fmt.Sprintf("%s\n\n%s\n", result, instructions)
}

func (screen *EditBankCardItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.CardNum, screen.newItemData.Expiry, screen.newItemData.CVV}

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
	screen.newItemData.CardNum = fields[2]
	screen.newItemData.Expiry = fields[3]
	screen.newItemData.CVV = fields[4]
}
