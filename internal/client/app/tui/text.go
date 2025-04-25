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

type ViewTextItemsScreen struct {
	options     []string
	category    Category
	itemManager *ItemManager
	backScreen  Screen
	list        *list.Model
}

type ViewTextDataScreen struct {
	backScreen Screen
	itemData   *textItemData
}

type textItemData struct {
	text string
}

type AddTextItemScreen struct {
	itemManager *ItemManager
	category    Category
	newTitle    string
	newDesc     string
	newItemData *textItemData
	createdTime string
	cursor      int
	backScreen  Screen
}

func (screen *ViewTextItemsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
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
				return ViewTextDataScreen{
					backScreen: screen,
					itemData: &textItemData{
						text: err.Error(),
					},
				}, nil
			}
			return ViewTextDataScreen{
				backScreen: screen,
				itemData: &textItemData{
					text: itemData.(*textItemData).text,
				},
			}, nil
		case "q":
			return screen.backScreen, nil // Go back when ESC is pressed
		}
	}
	return screen, nil
}

func (screen *ViewTextItemsScreen) View() string {
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

func (screen ViewTextDataScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return screen.backScreen, nil
		}
	}
	return screen, nil
}

func (screen ViewTextDataScreen) View() string {
	return screen.itemData.text
}

func (screen *AddTextItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
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
					//TODO create dedicated func
					resp, err := screen.itemManager.grpcClient.Handlers.TextHandler.PostTextData(context.Background(), &pb.PostTextDataRequest{
						Text:   screen.newItemData.text,
						TextId: "",
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
				if len(screen.newItemData.text) > 0 {
					screen.newItemData.text = screen.newItemData.text[:len(screen.newItemData.text)-1]
				}
			}

		case "ctrl+q": // Go back to the previous menu
			return screen.backScreen, nil
		case "up":
			screen.cursor = (screen.cursor - 1 + 3) % 3 // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % 3 // Focus on Description
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
				screen.newItemData.text += keyMsg.String()
			}
		}
	}
	return screen, nil
}

func (screen *AddTextItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &textItemData{}
	}
	var titleStyle, descStyle, text lipgloss.Style
	switch screen.cursor {
	case 0:
		titleStyle = selectedStyle // Highlight title when focused
		descStyle = unselectedStyle
		text = unselectedStyle

	case 1:
		titleStyle = unselectedStyle
		descStyle = selectedStyle // Highlight description when focused
		text = unselectedStyle
	case 2:
		titleStyle = unselectedStyle
		descStyle = unselectedStyle // Highlight description when focused
		text = selectedStyle
	}

	res := fmt.Sprintf("Add a new item:\n\n%s %s\n%s %s\n\n %s\n%s\n",
		titleStyle.Render("Title:"), titleStyle.Render(screen.newTitle),
		descStyle.Render("Description:"), descStyle.Render(screen.newDesc),
		text.Render("Text:"), text.Render(screen.newItemData.text),
	)

	res += separatorStyle.Render(fmt.Sprintf("Press Enter to save, ESC to cancel, or Backspace to delete the last character.\n"))

	return res
}
