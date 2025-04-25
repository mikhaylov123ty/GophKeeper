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

const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
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
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sText Information%s\n"+
			"=======================%s"+
			"%s%s\n",
		ColorBold, ColorReset,
		separator,
		ColorGreen, screen.itemData.text,
	)
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

		case "ctrl+q": // Go back to the previous menu
			return screen.backScreen, nil
		case "up":
			screen.cursor = (screen.cursor - 1 + 3) % 3 // Focus on Title
		case "down":
			screen.cursor = (screen.cursor + 1) % 3 // Focus on Description
		default:
			screen.handleInput(keyMsg.String())
		}

	}

	return screen, nil
}

func (screen *AddTextItemScreen) View() string {
	if screen.newItemData == nil {
		screen.newItemData = &textItemData{}
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
	addLine("Text:", screen.newItemData.text, styles[2])

	// Combine the lines with newlines
	result := strings.Join(lines, "\n")

	// Add instructions at the end
	instructions := "Press Enter to save, Q to cancel, or Backspace to delete the last character."

	return fmt.Sprintf("%s\n\n%s\n", result, instructions)
}

func (screen *AddTextItemScreen) handleInput(input string) {
	fields := []string{screen.newTitle, screen.newDesc, screen.newItemData.text}

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
	screen.newItemData.text = fields[2]
}
