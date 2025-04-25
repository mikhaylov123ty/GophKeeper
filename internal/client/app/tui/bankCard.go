package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type ViewBankCardItemsScreen struct {
	options     []string
	category    Category
	itemManager *ItemManager
	backScreen  Screen
	list        *list.Model
}

type ViewBankCardDataScreen struct {
	backScreen Screen
	itemData   *bankCardItemData
}

type bankCardItemData struct {
	cardNum string
	expiry  string
	cvv     string
}

type AddBankCardItemScreen struct {
	itemManager *ItemManager
	category    Category
	newTitle    string
	newDesc     string
	newItemData *bankCardItemData
	createdTime string // Set this to current time when item is created
	cursor      int    // 0 for Title, 1 for Description
	backScreen  Screen
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
			itemData, err := screen.itemManager.getItemData(itemDataID, screen.category)
			if err != nil {
				return ViewBankCardDataScreen{
					backScreen: screen,
					itemData: &bankCardItemData{
						cardNum: err.Error(),
					},
				}, nil
			}
			return ViewBankCardDataScreen{
				backScreen: screen,
				itemData: &bankCardItemData{
					cardNum: itemData.(*bankCardItemData).cardNum,
					expiry:  itemData.(*bankCardItemData).expiry,
					cvv:     itemData.(*bankCardItemData).cvv,
				},
			}, nil

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
		return "No items to display.\n\nPress ESC to go back.\n"
	}

	listItems := []list.Item{}
	for _, v := range screen.itemManager.metaItems[screen.category] {
		listItems = append(listItems, v)
	}

	screen.list.SetItems(listItems)

	return screen.list.View()
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
		ColorGreen, screen.itemData.cardNum, ColorReset,
		ColorYellow, screen.itemData.expiry, ColorReset,
		ColorRed, screen.itemData.cvv, ColorReset,
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
					//TODO maybe let server comnstruct metaD
					resp, err := screen.itemManager.grpcClient.Handlers.BankCardsHandler.PostBankCardData(context.Background(), &pb.PostBankCardDataRequest{
						CardNum: screen.newItemData.cardNum,
						Expiry:  screen.newItemData.expiry,
						Cvv:     screen.newItemData.cvv,
						CardId:  "",
						MetaData: &pb.MetaData{
							Id:          newItem.Id.String(),
							Title:       newItem.Title,
							Description: newItem.Description,
							DataType:    string(screen.category),
							UserId:      screen.itemManager.userID,
						},
					})
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
			screen.cursor = (screen.cursor - 1 + 5) % 5 // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % 5 // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}
	}
	return screen, nil
}

func (screen *AddBankCardItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &bankCardItemData{}
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
	addLine("Card Num:", screen.newItemData.cardNum, styles[2])
	addLine("Expiry:", screen.newItemData.expiry, styles[3])
	addLine("CVV:", screen.newItemData.cvv, styles[4])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	// Add instructions at the end
	instructions := "Press Enter to save, Q to cancel, or Backspace to delete the last character."

	return fmt.Sprintf("%s\n\n%s\n", result, instructions)
}

func (screen *AddBankCardItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.cardNum, screen.newItemData.expiry, screen.newItemData.cvv}

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
	screen.newItemData.cardNum = fields[2]
	screen.newItemData.expiry = fields[3]
	screen.newItemData.cvv = fields[4]
}

// safeDisplay ensures sensitive info is masked if necessary, e.g., hiding parts of the card number
func safeDisplay(cardNum string) string {
	if len(cardNum) > 4 {
		return "**** **** **** " + cardNum[len(cardNum)-4:] // Masking all but the last four digits
	}
	return cardNum // Return as-is if the length is less than or equal to 4
}
