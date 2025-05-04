package screens

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

const (
	cardFields = 5
)

// viewBankCardDataScreen represents a screen for viewing detailed bank card information.
// backScreen is the previous screen to return to upon user request.
// itemData holds the bank card data to be displayed on the screen.
type viewBankCardDataScreen struct {
	backScreen models.Screen
	itemData   *models.BankCardData
}

// addBankCardItemScreen represents a screen for adding or editing bank card information within the application.
// It embeds itemScreen to leverage shared functionality for managing items.
// The newItemData field holds the bank card details being added or edited.
type addBankCardItemScreen struct {
	*itemScreen
	newItemData *models.BankCardData
}

// Update processes the input message and determines the next screen state and command to execute.
func (screen *viewBankCardDataScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ:
			return screen.backScreen, nil
		}
	}

	return screen, nil
}

func (screen *viewBankCardDataScreen) View() string {
	body := utils.DataHeader()

	body += fmt.Sprintf(
		"\n%sCard Num: %s%s\n"+
			"%sExpiry: %s%s\n"+
			"%sCVV: %s%s\n",
		utils.ColorGreen, screen.itemData.CardNum, utils.ColorReset,
		utils.ColorYellow, screen.itemData.Expiry, utils.ColorReset,
		utils.ColorRed, screen.itemData.CVV, utils.ColorReset,
	)

	body += utils.ItemDataFooter()

	return body
}

func (screen *addBankCardItemScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			if screen.newTitle != "" && screen.newDesc != "" {
				if screen.newItemData != nil {
					cardData, err := json.Marshal(screen.newItemData)
					if err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}

					if err = screen.postItemData(cardData); err != nil {
						return &ErrorScreen{
							backScreen: screen,
							err:        err,
						}, nil
					}
				}
			}
			return screen.backScreen, nil // Go back to category menu
		case tea.KeyCtrlQ: // Go back to the previous menu
			return screen.backScreen, nil
		case tea.KeyUp:
			screen.cursor = (screen.cursor - 1 + cardFields) % cardFields // Focus on Title
		case tea.KeyDown:
			screen.cursor = (screen.cursor + 1) % cardFields // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}
	}
	return screen, nil
}

// TODO UNIFY THIS, like run build func in method
func (screen *addBankCardItemScreen) View() string {
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
	styles := []lipgloss.Style{
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
		utils.UnselectedStyle,
	}
	styles[screen.cursor] = utils.CursorStyle

	// Build each line
	addLine("Title:", screen.newTitle, styles[0])
	addLine("Description:", screen.newDesc, styles[1])
	addLine("Card Num:", screen.newItemData.CardNum, styles[2])
	addLine("Expiry:", screen.newItemData.Expiry, styles[3])
	addLine("CVV:", screen.newItemData.CVV, styles[4])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	result += utils.AddItemsFooter()

	return result
}

func (screen *addBankCardItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.CardNum, screen.newItemData.Expiry, screen.newItemData.CVV}

	// Backspace logic
	if input == "backspace" {
		if len(fields[screen.cursor]) > 0 {
			fields[screen.cursor] = fields[screen.cursor][:len(fields[screen.cursor])-1]
		}
	} else {
		// Ignore special keys
		if len(input) == 1 {
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
