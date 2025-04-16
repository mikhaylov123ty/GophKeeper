package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	currentScreen Screen
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
	managers   map[Category]*ItemManager
}

type ItemManager struct {
	items []string
}

type CategoryMenu struct {
	options     []string
	cursor      int
	itemManager *ItemManager
	backScreen  Screen
}

type ViewItemsScreen struct {
	itemManager *ItemManager
	backScreen  Screen
}

type AddItemScreen struct {
	itemManager *ItemManager
	newItem     string
	backScreen  Screen
}

type DeleteItemScreen struct {
	itemManager *ItemManager
	deleteIndex int
	backScreen  Screen
}

const (
	ViewOption   = "View items"
	AddOption    = "Add an item"
	DeleteOption = "Delete an item"
	BackOption   = "Back"
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
			return CategoryMenu{
				options:     []string{ViewOption, AddOption, DeleteOption, BackOption},
				cursor:      0,
				itemManager: m.managers[category],
				backScreen:  m,
			}, nil
		}
	}
	return m, nil
}

func (m MainMenu) View() string {
	s := "Main Menu:\n\n"
	for i, category := range m.categories {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, category)
	}
	s += "\nUse arrow keys to navigate and enter to select.\n"
	return s
}

func (cm CategoryMenu) Update(msg tea.Msg) (Screen, tea.Cmd) {
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
				return ViewItemsScreen{itemManager: cm.itemManager, backScreen: cm}, nil
			case 1: // Add item
				return AddItemScreen{itemManager: cm.itemManager, backScreen: cm}, nil
			case 2: // Delete item
				return DeleteItemScreen{itemManager: cm.itemManager, deleteIndex: -1, backScreen: cm}, nil
			case 3: // Back
				return cm.backScreen, nil
			}
		case "esc": // Go back to main menu
			return cm.backScreen, nil
		}
	}
	return cm, nil
}

func (cm CategoryMenu) View() string {
	s := "Category Menu:\n\n"
	for i, option := range cm.options {
		cursor := " "
		if cm.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}
	s += "\nPress ESC to go back to the main menu.\n"
	return s
}

func (screen ViewItemsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return screen.backScreen, nil // Go back when ESC is pressed
		}
		return screen.backScreen, nil // Default behavior to go back
	}
	return screen, nil
}

func (screen ViewItemsScreen) View() string {
	if len(screen.itemManager.items) == 0 {
		return "No items to display.\n\nPress ESC to go back.\n"
	}

	s := "Items:\n\n"
	for i, item := range screen.itemManager.items {
		s += fmt.Sprintf("%d: %s\n", i+1, item)
	}
	s += "\nPress ESC to go back.\n"
	return s
}

func (screen AddItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if screen.newItem != "" {
				screen.itemManager.items = append(screen.itemManager.items, screen.newItem)
			}
			return screen.backScreen, nil // Go back to category menu
		case "backspace":
			if len(screen.newItem) > 0 {
				screen.newItem = screen.newItem[:len(screen.newItem)-1]
			}
		case "esc": // Go back to the previous menu
			return screen.backScreen, nil
		default:
			screen.newItem += keyMsg.String()
		}
	}
	return screen, nil
}

func (screen AddItemScreen) View() string {
	return fmt.Sprintf("Add a new item:\n\n%s\n\nPress Enter to save, ESC to cancel, or Backspace to delete the last character.\n", screen.newItem)
}

func (screen DeleteItemScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if screen.deleteIndex >= 0 && screen.deleteIndex < len(screen.itemManager.items) {
				// Delete the selected item
				screen.itemManager.items = append(screen.itemManager.items[:screen.deleteIndex], screen.itemManager.items[screen.deleteIndex+1:]...)
			}
			return screen.backScreen, nil // Go back to category menu
		case "backspace":
			if screen.deleteIndex >= 1 {
				screen.deleteIndex--
			}
		case "esc": // Go back to the previous menu
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
	for i, item := range screen.itemManager.items {
		s += fmt.Sprintf("%d: %s\n", i+1, item)
	}
	s += "\nSelect the item number to delete (Press ESC to go back):\n"
	return s
}

func NewItemManager() *Model {
	itemManagers := map[Category]*ItemManager{
		TextCategory: {items: []string{}},
		FileCategory: {items: []string{}},
		CardCategory: {items: []string{}},
	}

	model := Model{
		currentScreen: MainMenu{
			categories: []Category{TextCategory, FileCategory, CardCategory, ExitCategory}, // Include exit category
			cursor:     0,
			managers:   itemManagers,
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
