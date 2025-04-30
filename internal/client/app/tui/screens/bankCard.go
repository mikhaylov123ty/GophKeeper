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
	cardFields = 5
)

type viewBankCardDataScreen struct {
	backScreen models.Screen
	itemData   *dbModels.BankCardData
}

type addBankCardItemScreen struct {
	*models.ItemScreen
	selectedItem *models.MetaItem
	newItemData  *dbModels.BankCardData
}

func (screen *viewBankCardDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			return screen.backScreen, nil
		}
	}

	return screen, nil
}

// TODO UNIFY THIS, like run build func in method
func (screen *viewBankCardDataScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	body := fmt.Sprintf(
		"%sCard Information%s\n"+
			"=======================%s"+
			"%sCard Num: %s%s\n"+
			"%sExpiry: %s%s\n"+
			"%sCVV: %s%s\n",
		utils.ColorBold, utils.ColorReset, separator,
		utils.ColorGreen, screen.itemData.CardNum, utils.ColorReset,
		utils.ColorYellow, screen.itemData.Expiry, utils.ColorReset,
		utils.ColorRed, screen.itemData.CVV, utils.ColorReset,
	)
	body += utils.ItemDataFooter()

	return body
}

func (screen *addBankCardItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.NewTitle != "" && screen.NewDesc != "" {
				if screen.newItemData != nil {
					cardData, err := json.Marshal(screen.newItemData)
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

					resp, err := screen.ItemsManager.PostItemData(cardData, dataID, &metaData)
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
			return screen.BackScreen, nil // Go back to category menu

		case "ctrl+q": // Go back to the previous menu
			return screen.BackScreen, nil
		case "up":
			screen.Cursor = (screen.Cursor - 1 + cardFields) % cardFields // Focus on Title
		case "down":
			screen.Cursor = (screen.Cursor + 1) % cardFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}
	}
	return screen, nil
}

// TODO UNIFY THIS, like run build func in method
func (screen *addBankCardItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &dbModels.BankCardData{}
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
		utils.UnselectedStyle,
	}
	styles[screen.Cursor] = utils.SelectedStyle

	// Build each line
	addLine("Title:", screen.NewTitle, styles[0])
	addLine("Description:", screen.NewDesc, styles[1])
	addLine("Card Num:", screen.newItemData.CardNum, styles[2])
	addLine("Expiry:", screen.newItemData.Expiry, styles[3])
	addLine("CVV:", screen.newItemData.CVV, styles[4])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addBankCardItemScreen) handleInput(input string) {
	fields := []string{screen.NewTitle, screen.NewDesc, screen.newItemData.CardNum, screen.newItemData.Expiry, screen.newItemData.CVV}

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
	screen.newItemData.CardNum = fields[2]
	screen.newItemData.Expiry = fields[3]
	screen.newItemData.CVV = fields[4]
}
