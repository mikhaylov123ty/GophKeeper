package tui

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"strconv"
	"time"
)

type Model struct {
	currentScreen Screen
	grpcClient    *grpc.Client
}

type Screen interface {
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
}

type Category string

const (
	TextCategory Category = "Text"
	FileCategory Category = "Files"
	CardCategory Category = "Cards"
	ExitCategory Category = "Exit" // New exit category

)

type MainMenu struct {
	categories []Category
	cursor     int
	manager    *ItemManager
	nextScreen Screen
}

type ActionsMenu struct {
	options     []string
	cursor      int
	category    Category
	itemManager *ItemManager
	backScreen  Screen
}

type ItemManager struct {
	metaItems   map[Category][]*MetaItem
	textHandler pb.TextHandlersClient
}

func (im *ItemManager) getItemData(dataID string, category Category) (any, error) {
	switch category {
	case TextCategory:
		response, err := im.textHandler.GetTextData(context.Background(), &pb.GetTextDataRequest{
			TextId: dataID,
		})
		if err != nil {
			return nil, err
		}
		return &textItemData{
			text: response.GetText(),
		}, nil
	}
	return nil, nil
}

type MetaItem struct {
	Num         int
	Id          uuid.UUID
	Title       string
	dataID      string
	Description string
	Created     string // You can use time.Time for actual timestamp
	Modified    string // Same as above
}

type ViewTextItemsScreen struct {
	options     []string
	cursor      int
	category    Category
	itemManager *ItemManager
	backScreen  Screen
}

type ViewTextDataScreen struct {
	backScreen Screen
	itemData   string
}

type ViewBankCardItemsScreen struct {
	options     []string
	cursor      int
	category    Category
	itemManager *ItemManager
	backScreen  Screen
}

type AddTextItemScreen struct {
	itemManager *ItemManager
	category    Category
	newTitle    string
	newDesc     string
	newItemData *textItemData
	createdTime string // Set this to current time when item is created
	cursor      int    // 0 for Title, 1 for Description
	backScreen  Screen
}

type textItemData struct {
	text string
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

type bankCardItemData struct {
	cardNum string
	expiry  time.Time
	cvv     string
}

type DeleteItemScreen struct {
	itemManager *ItemManager
	category    Category
	deleteIndex int
	backScreen  Screen
}

const (
	ViewOption   = "View items"
	AddOption    = "Add an item"
	DeleteOption = "Delete an item"
	BackOption   = "Back"
)

// Style definitions
var (
	cursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))                     // Bright purple
	selectedStyle   = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("2")).Bold(true) // Green
	unselectedStyle = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("7"))            // White
	backgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("235"))                     // Grey background
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("65"))           // Bold yellow
	separatorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))                     // Light grey
)

func (m MainMenu) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.cursor = (m.cursor + 1) % len(m.categories)
		case "up":
			m.cursor = (m.cursor - 1 + len(m.categories)) % len(m.categories)
		case "enter":
			category := m.categories[m.cursor]
			if category == ExitCategory {
				return m, tea.Quit // Exit the application
			}
			m.nextScreen = &ActionsMenu{
				options:     []string{ViewOption, AddOption, DeleteOption, BackOption},
				category:    category,
				itemManager: m.manager,
				backScreen:  m,
			}
			return m.nextScreen, nil
		}
	}

	return m, nil
}

func (m MainMenu) View() string {
	s := titleStyle.Render("Main Menu:\n\n")
	for i, category := range m.categories {
		if m.cursor == i {
			s += selectedStyle.Render(fmt.Sprintf("[x] %s\n", string(category))) // Selected option with color
		} else {
			s += unselectedStyle.Render(fmt.Sprintf("[ ] %s\n", category)) // Unselected option with color
		}
	}
	s += separatorStyle.Render("Use arrow keys to navigate and enter to select.\n") // Navigation instructions
	return s
}

func (cm ActionsMenu) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			cm.cursor = (cm.cursor + 1) % len(cm.options)
		case "up":
			cm.cursor = (cm.cursor - 1 + len(cm.options)) % len(cm.options)
		case "enter":
			switch cm.cursor {
			case 0: // View items
				switch cm.category {
				case TextCategory:
					return ViewTextItemsScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case CardCategory:
					return ViewBankCardItemsScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				}
			case 1: // Add item
				switch cm.category {
				case TextCategory:
					return &AddTextItemScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				case CardCategory:
					return AddBankCardItemScreen{itemManager: cm.itemManager, backScreen: cm, category: cm.category}, nil
				}
			case 2: // Delete item
				return DeleteItemScreen{itemManager: cm.itemManager, deleteIndex: -1, backScreen: cm, category: cm.category}, nil
			case 3: // Back
				return cm.backScreen, nil
			}
		case "esc": // Go back to main menu
			return cm.backScreen, nil
		}
	}
	return cm, nil
}

func (cm ActionsMenu) View() string {
	s := titleStyle.Render("Category Menu:\n\n")
	for i, option := range cm.options {
		if cm.cursor == i {
			s += selectedStyle.Render(fmt.Sprintf("[x] %s\n", option)) // Selected option with color
		} else {
			s += unselectedStyle.Render(fmt.Sprintf("[ ] %s\n", option)) // Unselected option with color
		}
	}

	s += separatorStyle.Render("Press ESC to go back to the main menu.\n") // Navigation instructions
	return s
}

func (screen ViewTextItemsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if len(screen.itemManager.metaItems[screen.category]) > 0 {
				screen.cursor = (screen.cursor + 1) % len(screen.itemManager.metaItems[screen.category])
			}
		case "up":
			if len(screen.itemManager.metaItems[screen.category]) > 0 {
				screen.cursor = (screen.cursor - 1 + len(screen.itemManager.metaItems[screen.category])) % len(screen.itemManager.metaItems[screen.category])
			}
		case "enter":
			itemData, err := screen.itemManager.getItemData(screen.itemManager.metaItems[screen.category][screen.cursor].dataID, screen.category)
			if err != nil {
				return ViewTextDataScreen{
					backScreen: screen,
					itemData:   err.Error(),
				}, nil
			}
			return ViewTextDataScreen{
				backScreen: screen,
				itemData:   itemData.(*textItemData).text,
			}, nil
		case "q":
			return screen.backScreen, nil // Go back when ESC is pressed
		}
	}
	return screen, nil
}

func (screen ViewTextItemsScreen) View() string {
	if len(screen.itemManager.metaItems[screen.category]) == 0 {
		return cursorStyle.Render("No items to display.\n\nPress ESC to go back.\n")
	}
	s := "Items:\n\n"
	for i, item := range screen.itemManager.metaItems[screen.category] {
		if screen.cursor == i {
			s += selectedStyle.Render(fmt.Sprintf("[x] %d.\n Title: %s\n    Description: %s\n    Created: %s\n    Modified: %s\n",
				item.Num, item.Title, item.Description, item.Created, item.Modified)) // Display item details
		} else {
			s += unselectedStyle.Render(fmt.Sprintf("[ ] %d.\n Title: %s\n    Description: %s\n    Created: %s\n    Modified: %s\n",
				item.Num, item.Title, item.Description, item.Created, item.Modified))
		}
	}

	s += separatorStyle.Render("Press ESC to go back.\n") // Navigation instructions
	return s
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
	return screen.itemData
}

func (screen ViewBankCardItemsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "down":
			screen.cursor = (screen.cursor + 1) % len(screen.itemManager.metaItems[screen.category])
		case "up":
			screen.cursor = (screen.cursor - 1 + len(screen.itemManager.metaItems[screen.category])) % len(screen.itemManager.metaItems[screen.category])
		//case "enter":
		//
		//	item := screen.itemManager.metaItems[screen.category][screen.cursor]
		//	screen.itemManager.GetItemData(screen.category, item.dataID)

		case "q":
			return screen.backScreen, nil // Go back when ESC is pressed
		}
	}
	return screen, nil
}

func (screen ViewBankCardItemsScreen) View() string {
	if len(screen.itemManager.metaItems[screen.category]) == 0 {
		return "No items to display.\n\nPress ESC to go back.\n"
	}

	s := "Items:\n\n"
	for i, item := range screen.itemManager.metaItems[screen.category] {
		if screen.cursor == i {
			s += selectedStyle.Render(fmt.Sprintf("[x] %d.\n Title: %s\n    Description: %s\n    Created: %s\n    Modified: %s\n",
				item.Num, item.Title, item.Description, item.Created, item.Modified)) // Display item details
		} else {
			s += unselectedStyle.Render(fmt.Sprintf("[ ] %d.\n Title: %s\n    Description: %s\n    Created: %s\n    Modified: %s\n",
				item.Num, item.Title, item.Description, item.Created, item.Modified))
		}
	}

	s += separatorStyle.Render("Press ESC to go back.\n") // Navigation instructions
	return s
}

func (screen *AddTextItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" {
				// Create new item and add to the manager
				newItem := MetaItem{
					Num:         len(screen.itemManager.metaItems[screen.category]) + 1, // Set item number
					Id:          uuid.New(),
					Title:       screen.newTitle,
					Description: screen.newDesc,
				}

				if screen.newItemData != nil {

					//TODO maybe let server comnstruct metaD
					resp, err := screen.itemManager.textHandler.PostTextData(context.Background(), &pb.PostTextDataRequest{
						Text:   screen.newItemData.text,
						TextId: "",
						MetaData: &pb.MetaData{
							Id:          newItem.Id.String(),
							Title:       newItem.Title,
							Description: newItem.Description,
							DataType:    string(screen.category),
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

func (screen AddBankCardItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newTitle != "" && screen.newDesc != "" {
				// Create new item and add to the manager
				newItem := MetaItem{
					Num:         len(screen.itemManager.metaItems[screen.category]) + 1, // Set item number
					Title:       screen.newTitle,
					Description: screen.newDesc,
					Created:     time.Now().Format(time.RFC3339), // Current time
					Modified:    time.Now().Format(time.RFC3339), // Current time
				}

				if screen.newItemData != nil {
					//todo send itemdata to server, get id, complete meta and send meta
					screen.itemManager.metaItems[screen.category] = append(screen.itemManager.metaItems[screen.category], &newItem)
				}

			}
			return screen.backScreen, nil // Go back to category menu
		case "backspace":
			if len(screen.newTitle) > 0 {
				screen.newTitle = screen.newTitle[:len(screen.newTitle)-1]
			}
		case "ctrl+q": // Go back to the previous menu
			return screen.backScreen, nil
			// TODO roll
		case "up":
			screen.cursor = 0 // Focus on Title
		case "down":
			screen.cursor = 1 // Focus on Description
		}
		// Handle character inputs depending on the focused field
		if screen.cursor == 0 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" { // Ignore special keys
				screen.newTitle += keyMsg.String()
			}
		} else if screen.cursor == 1 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" { // Ignore special keys
				screen.newDesc += keyMsg.String()
			}
		} else if screen.cursor == 2 {
			if keyMsg.String() != "up" && keyMsg.String() != "down" && keyMsg.String() != "esc" {
				screen.newItemData.cardNum += keyMsg.String()
			}
		}
	}
	return screen, nil
}

func (screen AddBankCardItemScreen) View() string {
	var titleStyle, descStyle, text lipgloss.Style
	screen.newItemData = &bankCardItemData{}

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

	res := fmt.Sprintf("Add a new item:\n\n%s %s\n%s %s\n\n %s\n%s\n %s\n%s\n %s\n%s\n",
		titleStyle.Render("Title:"), titleStyle.Render(screen.newTitle),
		descStyle.Render("Description:"), descStyle.Render(screen.newDesc),
		text.Render("Card Num:"), descStyle.Render(screen.newItemData.cardNum),
		text.Render("Expiry:"), descStyle.Render(screen.newItemData.expiry.String()),
		text.Render("CVV:"), descStyle.Render(screen.newItemData.cvv),
	)

	return fmt.Sprintf("%s\n\nPress Enter to save, ESC to cancel, or Backspace to delete the last character.\n", res)
}

func (screen DeleteItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if screen.deleteIndex >= 0 && screen.deleteIndex < len(screen.itemManager.metaItems[screen.category]) {
				// Delete the selected item
				screen.itemManager.metaItems[screen.category] = append(screen.itemManager.metaItems[screen.category][:screen.deleteIndex], screen.itemManager.metaItems[screen.category][screen.deleteIndex+1:]...)
			}
			return screen.backScreen, nil // Go back to category menu
		case "backspace":
			if screen.deleteIndex >= 1 {
				screen.deleteIndex--
			}
		case "q": // Go back to the previous menu
			return screen.backScreen, nil
		default:
			if key, err := strconv.Atoi(msg.String()); err == nil && key > 0 {
				screen.deleteIndex = key - 1 // Considering 1-based user input
			}
		}
	}
	return screen, nil
}

func (screen DeleteItemScreen) View() string {
	s := "Delete an item:\n\n"
	for i, item := range screen.itemManager.metaItems[screen.category] {
		s += fmt.Sprintf("  - %d: %v\n", i+1, item) // Use bullet points for items
	}
	s += "Select the item number to delete (Press ESC to go back):\n" // Navigation instructions
	return s
}

func NewItemManager(grpcClient *grpc.Client) *Model {
	model := Model{
		currentScreen: MainMenu{
			categories: []Category{TextCategory, FileCategory, CardCategory, ExitCategory}, // Include exit category
			cursor:     0,
			manager: &ItemManager{
				metaItems:   map[Category][]*MetaItem{},
				textHandler: grpcClient.TextHandler,
			},
		},
	}
	return &model
}

func (m Model) Init() tea.Cmd {
	return nil // No command to run, initial screen is set.
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	nextScreen, cmd := m.currentScreen.Update(msg)
	m.currentScreen = nextScreen
	return m, cmd
}

func (m Model) View() string {
	return m.currentScreen.View()
}
