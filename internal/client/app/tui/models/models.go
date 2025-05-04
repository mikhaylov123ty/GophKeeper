package models

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

// Screen is an interface defining methods for screen management used in a terminal-based UI application.
// Update handles messages or events and returns the updated Screen along with an optional command to execute.
// View returns the string representation of the current screen for rendering.
type Screen interface {
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
}

// ItemsManager provides methods for managing items and metadata efficiently.
// GetMetaData retrieves metadata items associated with the provided string key.
// SaveMetaItem stores a metadata item under the specified string key.
// PostItemData records item data with provided data, string key, and metadata.
// GetItemData fetches item data associated with the given string key.
// DeleteItem removes an item using uuid, string key, and additional parameters.
// PostUserData handles user authentication by posting user data with the given credentials.
// SyncMeta synchronizes the metadata across the system.
type ItemsManager interface {
	GetMetaData(string) []*MetaItem
	SaveMetaItem(string, *MetaItem)
	PostItemData([]byte, string, *pb.MetaData) (*pb.PostItemDataResponse, error)
	GetItemData(string) (string, error)
	DeleteItem(uuid.UUID, string, string) error
	PostUserData(string, string) error
	SyncMeta() error
}

// CredsData represents a structure containing login credentials with a Login and Password field.
type CredsData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// TextData represents a structure for storing a single string-based textual data field.
type TextData struct {
	Text string `json:"text"`
}

// BankCardData represents the structure for storing bank card details
// including card number, expiry date, and CVV.
type BankCardData struct {
	CardNum string `json:"card_num"`
	Expiry  string `json:"expiry"`
	CVV     string `json:"cvv"`
}

// Binary represents a file with its metadata and path information stored as structured data.
type Binary struct {
	Binary   BinaryData `json:"binary_data"`
	FilePath string     `json:"name"`
}

// BinaryData represents a binary file with its metadata and content.
// It includes the file name, type, data, and size.
type BinaryData struct {
	Name     string `json:"name"`
	Content  []byte `json:"content"`
	FileSize int    `json:"file_size"`
}

// Model represents the primary application state,
// managing screen transitions within the terminal UI.
type Model struct {
	CurrentScreen Screen
}

// Init initializes the program's starting state and
// does not trigger any command execution.
func (m Model) Init() tea.Cmd {
	return nil // No command to run, initial screen is set.
}

// Update processes incoming messages, updates the current screen state,
// and returns an updated model and optional command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	nextScreen, cmd := m.CurrentScreen.Update(msg)
	m.CurrentScreen = nextScreen
	return m, cmd
}

// View returns the string representation of the current screen for rendering.
func (m Model) View() string {
	return m.CurrentScreen.View()
}

// MetaItem represents metadata associated with an item,
// including its ID, title, description, and timestamps.
type MetaItem struct {
	ID          uuid.UUID
	Title       string
	Description string
	DataID      string
	Created     string
	Modified    string
}

// FilterValue returns a string representation used to filter or identify the MetaItem.
// Currently, it returns an empty string.
func (m MetaItem) FilterValue() string { return "" }

// MetaItemDelegate manages the rendering, spacing,
// and height for a list of MetaItem instances in the UI.
type MetaItemDelegate struct{}

// Height returns the constant height of a MetaItemDelegate,
// which is used to define the height of each list item.
func (d MetaItemDelegate) Height() int { return 1 }

// Spacing returns the space between list items in the MetaItemDelegate, measured in lines.
// It is set to a default of 0.
func (d MetaItemDelegate) Spacing() int { return 0 }

// Update handles messages and updates the state of the MetaItemDelegate.
// No actions are performed in the current implementation.
func (d MetaItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders a MetaItem within a list, applying specific styles
// for selected and non-selected items.
func (d MetaItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	itemStyle := lipgloss.NewStyle().PaddingLeft(4)
	selecteditemStyle := utils.CursorStyle.PaddingLeft(2)
	i, ok := listItem.(*MetaItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. Title: %s | Description: %s | Created: %s | Modified: %s\n", index+1, i.Title, i.Description, i.Created, i.Modified)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selecteditemStyle.Render("[x] " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
