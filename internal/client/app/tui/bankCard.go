package tui

import (
	"context"
	"fmt"

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

	return fmt.Sprintf("Card Num: %s\nExpiry: %s\nCVV: %s\n", screen.itemData.cardNum, screen.itemData.expiry, screen.itemData.cvv)
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
		case "backspace":
			switch screen.cursor {
			case 0:
				if len(screen.newTitle) > 0 {
					screen.newTitle = screen.newTitle[:len(screen.newTitle)-1]
				}
			case 1:
				if len(screen.newDesc) > 0 {
					screen.newDesc = screen.newDesc[:len(screen.newDesc)-1]
				}
			case 2:
				if len(screen.newItemData.cardNum) > 0 {
					screen.newItemData.cardNum = screen.newItemData.cardNum[:len(screen.newItemData.cardNum)-1]
				}
			case 3:
				if len(screen.newItemData.expiry) > 0 {
					screen.newItemData.expiry = screen.newItemData.expiry[:len(screen.newItemData.expiry)-1]
				}
			case 4:
				if len(screen.newItemData.cvv) > 0 {
					screen.newItemData.cvv = screen.newItemData.cvv[:len(screen.newItemData.cvv)-1]
				}
			}

		case "ctrl+q": // Go back to the previous menu
			return screen.backScreen, nil
		case "up":
			screen.cursor = (screen.cursor - 1 + 5) % 5 // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % 5 // Focus on Description
		}

		// Handle character inputs depending on the focused field
		// TODO cursor item to var and use it in operations such these and other buttons
		if screen.cursor == 0 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" && keyMsg.String() != "backspace" { // Ignore special keys
				screen.newTitle += keyMsg.String()
			}
		} else if screen.cursor == 1 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" && keyMsg.String() != "backspace" { // Ignore special keys
				screen.newDesc += keyMsg.String()
			}
		} else if screen.cursor == 2 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" && keyMsg.String() != "backspace" {
				screen.newItemData.cardNum += keyMsg.String()
			}

		} else if screen.cursor == 3 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" && keyMsg.String() != "backspace" {
				screen.newItemData.expiry += keyMsg.String()
			}
		} else if screen.cursor == 4 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" && keyMsg.String() != "backspace" {
				screen.newItemData.cvv += keyMsg.String()
			}
		}
	}
	return screen, nil
}

func (screen *AddBankCardItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &bankCardItemData{}
	}

	var titleStyle, descStyle, cardNum, expiry, cvv lipgloss.Style

	switch screen.cursor {
	case 0:
		titleStyle = selectedStyle // Highlight title when focused
		descStyle = unselectedStyle
		cardNum = unselectedStyle
		expiry = unselectedStyle
		cvv = unselectedStyle

	case 1:
		titleStyle = unselectedStyle
		descStyle = selectedStyle // Highlight description when focused
		cardNum = unselectedStyle
		expiry = unselectedStyle
		cvv = unselectedStyle
	case 2:
		titleStyle = unselectedStyle
		descStyle = unselectedStyle // Highlight description when focused
		cardNum = selectedStyle
		expiry = unselectedStyle
		cvv = unselectedStyle
	case 3:
		titleStyle = unselectedStyle
		descStyle = unselectedStyle // Highlight description when focused
		cardNum = unselectedStyle
		expiry = selectedStyle
		cvv = unselectedStyle
	case 4:
		titleStyle = unselectedStyle
		descStyle = unselectedStyle // Highlight description when focused
		cardNum = unselectedStyle
		expiry = unselectedStyle
		cvv = selectedStyle
	}

	res := fmt.Sprintf("Add a new item:\n\n%s %s\n%s %s\n\n %s\n%s\n %s\n%s\n %s\n%s\n",
		titleStyle.Render("Title:"), titleStyle.Render(screen.newTitle),
		descStyle.Render("Description:"), descStyle.Render(screen.newDesc),
		cardNum.Render("Card Num:"), cardNum.Render(screen.newItemData.cardNum),
		expiry.Render("Expiry:"), expiry.Render(screen.newItemData.expiry),
		cvv.Render("CVV:"), cvv.Render(screen.newItemData.cvv),
	)

	return fmt.Sprintf("%s\n\nPress Enter to save, ESC to cancel, or Backspace to delete the last character.\n", res)
}
