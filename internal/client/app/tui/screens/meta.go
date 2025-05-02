package screens

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	dbModels "github.com/mikhaylov123ty/GophKeeper/internal/models"
)

type viewMetaItemsScreen struct {
	category     string
	itemsManager models.ItemsManager
	backScreen   models.Screen
	list         *list.Model
	listTitle    string
}

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

// TODO make common unmarhsaler
func (screen *viewMetaItemsScreen) routeViewData(itemData string, category string) models.Screen {
	switch category {
	case TextCategory:
		var textData dbModels.TextData
		if err := json.Unmarshal([]byte(itemData), &textData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewTextDataScreen{
			backScreen: screen,
			itemData: &dbModels.TextData{
				Text: textData.Text,
			},
		}

	case CardCategory:
		var cardData dbModels.BankCardData
		if err := json.Unmarshal([]byte(itemData), &cardData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewBankCardDataScreen{
			backScreen: screen,
			itemData: &dbModels.BankCardData{
				CardNum: cardData.CardNum,
				Expiry:  cardData.Expiry,
				CVV:     cardData.CVV,
			},
		}

	case CredsCategory:
		var credsData dbModels.CredsData
		if err := json.Unmarshal([]byte(itemData), &credsData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewCredsDataScreen{
			backScreen: screen,
			itemData: &dbModels.CredsData{
				Login:    credsData.Login,
				Password: credsData.Password,
			},
		}

	case FileCategory:
		var binaryData dbModels.BinaryData
		if err := json.Unmarshal([]byte(itemData), &binaryData); err != nil {
			return &ErrorScreen{
				backScreen: screen,
				err:        err,
			}
		}

		return &viewBinaryDataScreen{
			backScreen: screen,
			itemData: &dbModels.BinaryData{
				Name:        binaryData.Name,
				ContentType: binaryData.ContentType,
				Content:     binaryData.Content,
			},
		}
	}

	return screen
}

func (screen *viewMetaItemsScreen) routeEditData(category string) models.Screen {
	selectedItem, ok := screen.list.SelectedItem().(*models.MetaItem)
	if !ok || selectedItem == nil {
		return &ErrorScreen{backScreen: screen, err: fmt.Errorf("item not found")}
	}

	switch category {
	case TextCategory:
		return &addTextItemScreen{
			itemScreen: &itemScreen{
				itemsManager: screen.itemsManager,
				backScreen:   screen,
				category:     screen.category,
				newTitle:     selectedItem.Title,
				newDesc:      selectedItem.Description,
				selectedItem: selectedItem,
			},
		}

	case CardCategory:
		return &addBankCardItemScreen{
			itemScreen: &itemScreen{
				itemsManager: screen.itemsManager,
				backScreen:   screen,
				category:     screen.category,
				newTitle:     selectedItem.Title,
				newDesc:      selectedItem.Description,
				selectedItem: selectedItem,
			},
		}

	case CredsCategory:
		return &addCredsItemScreen{
			itemScreen: &itemScreen{
				itemsManager: screen.itemsManager,
				backScreen:   screen,
				category:     screen.category,
				newTitle:     selectedItem.Title,
				newDesc:      selectedItem.Description,
			},
		}

	case FileCategory:
		return &addBinaryItemScreen{
			itemScreen: &itemScreen{
				itemsManager: screen.itemsManager,
				backScreen:   screen,
				category:     screen.category,
				newTitle:     selectedItem.Title,
				newDesc:      selectedItem.Description,
				selectedItem: selectedItem,
			},
		}
	}

	return screen
}
