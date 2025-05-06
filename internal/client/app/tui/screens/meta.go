package screens

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

// viewMetaItemsScreen represents a screen for viewing metadata items of a particular category.
// category specifies the metadata category to be viewed.
// itemsManager provides the methods required to manage metadata items.
// backScreen defines the previous screen for handling navigation.
// list is the UI model for displaying and interacting with metadata items.
// listTitle represents the title displayed at the top of the list.
type viewMetaItemsScreen struct {
	category     string
	itemsManager models.ItemsManager
	backScreen   models.Screen
	list         *list.Model
	listTitle    string
}

// View generates and returns the string representation of the viewMetaItemsScreen for rendering in the UI.
func (screen *viewMetaItemsScreen) View() string {
	if screen.list == nil {
		listModel := list.New([]list.Item{}, models.MetaItemDelegate{}, 10, utils.ListHeight)
		screen.list = &listModel
	}

	metaData := screen.itemsManager.GetMetaData(screen.category)

	if len(metaData) == 0 {
		return utils.SelectedStyle.Render("No items to display.\n\n") + utils.ItemDataFooter()
	}

	listItems := []list.Item{}
	for _, v := range metaData {
		listItems = append(listItems, v)
	}

	screen.list.SetItems(listItems)
	screen.list.SetShowHelp(false)
	screen.list.Title = screen.listTitle + " List"

	s := screen.list.View()
	s += utils.ListItemsFooter()

	return s
}

// Update handles user input and updates the state of the viewMetaItemsScreen based on the received message.
func (screen *viewMetaItemsScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			screen.list.CursorDown()

		case "up":
			screen.list.CursorUp()

		case "enter":
			if screen.list.SelectedItem() == nil {
				return screen.backScreen, nil
			}
			itemDataID := screen.list.SelectedItem().(*models.MetaItem).DataID
			itemData, err := screen.itemsManager.GetItemData(itemDataID)
			if err != nil {
				return &ErrorScreen{
					backScreen: screen,
					err:        err,
				}, nil
			}

			return screen.routeViewData(itemData, screen.category), nil

		case "e":
			if screen.list.SelectedItem() != nil {
				return screen.routeEditData(screen.category), nil
			}

		case "d":
			if len(screen.list.Items()) > 0 {
				if err := screen.itemsManager.DeleteItem(
					screen.list.SelectedItem().(*models.MetaItem).ID,
					screen.category,
					screen.list.SelectedItem().(*models.MetaItem).DataID,
				); err != nil {
					return &ErrorScreen{
						backScreen: screen,
						err:        err,
					}, nil
				}
				screen.list.CursorUp()
			}

		case "ctrl+q":
			return screen.backScreen, nil // Go back when ESC is pressed
		}
	}
	return screen, nil
}

// routeViewData maps the provided item data and category to the corresponding screen type for detailed view rendering.
// Returns a specific data screen or error screen if the data cannot be unmarshalled or processed.
// TODO build common unmarshaler func
func (screen *viewMetaItemsScreen) routeViewData(itemData string, category string) models.Screen {
	switch category {
	case TextCategory:
		var textData models.TextData
		if err := json.Unmarshal([]byte(itemData), &textData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewTextDataScreen{
			backScreen: screen,
			itemData: &models.TextData{
				Text: textData.Text,
			},
		}

	case CardCategory:
		var cardData models.BankCardData
		if err := json.Unmarshal([]byte(itemData), &cardData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewBankCardDataScreen{
			backScreen: screen,
			itemData: &models.BankCardData{
				CardNum: cardData.CardNum,
				Expiry:  cardData.Expiry,
				CVV:     cardData.CVV,
			},
		}

	case CredsCategory:
		var credsData models.CredsData
		if err := json.Unmarshal([]byte(itemData), &credsData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewCredsDataScreen{
			backScreen: screen,
			itemData: &models.CredsData{
				Login:    credsData.Login,
				Password: credsData.Password,
			},
		}

	case FileCategory:
		var binaryData models.BinaryData
		if err := json.Unmarshal([]byte(itemData), &binaryData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewBinaryDataScreen{
			backScreen: screen,
			itemData: &models.BinaryData{
				Name:     binaryData.Name,
				Content:  binaryData.Content,
				FileSize: binaryData.FileSize,
			},
		}
	}

	return screen
}

// routeEditData routes to the appropriate screen for editing an item based on its category and selected metadata item.
// Returns an editing screen for the given category or ErrorScreen if the item is not found or invalid.
func (screen *viewMetaItemsScreen) routeEditData(category string) models.Screen {
	selectedItem, ok := screen.list.SelectedItem().(*models.MetaItem)
	if !ok || selectedItem == nil {
		return &ErrorScreen{backScreen: screen, err: fmt.Errorf("item not found")}
	}

	createItemScreen := func() *itemScreen {
		return &itemScreen{
			itemsManager: screen.itemsManager,
			backScreen:   screen,
			category:     screen.category,
			newTitle:     selectedItem.Title,
			newDesc:      selectedItem.Description,
			selectedItem: selectedItem,
		}
	}

	// Map category to corresponding screen
	switch category {
	case TextCategory:
		return &addTextItemScreen{itemScreen: createItemScreen()}
	case CardCategory:
		return &addBankCardItemScreen{itemScreen: createItemScreen()}
	case CredsCategory:
		return &addCredsItemScreen{itemScreen: createItemScreen()}
	case FileCategory:
		return &addBinaryItemScreen{itemScreen: createItemScreen()}
	}

	return screen
}
