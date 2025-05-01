package models

import (
	"fmt"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type Screen interface {
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
}

type ItemsManager interface {
	GetMetaData(string) []*MetaItem
	SaveMetaItem(string, *MetaItem)
	PostItemData([]byte, string, *pb.MetaData) (*pb.PostItemDataResponse, error)
	GetItemData(string) (string, error)
	DeleteItem(uuid.UUID, string, string) error
	PostUserData(string, string) error
	SyncMeta() error
}

type Model struct {
	CurrentScreen Screen
}

func (m Model) Init() tea.Cmd {
	return nil // No command to run, initial screen is set.
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	nextScreen, cmd := m.CurrentScreen.Update(msg)
	m.CurrentScreen = nextScreen
	return m, cmd
}

func (m Model) View() string {
	return m.CurrentScreen.View()
}

type MetaItem struct {
	Id          uuid.UUID
	Title       string
	Description string
	DataID      string
	Created     string
	Modified    string
}

func (m MetaItem) FilterValue() string { return "" }

type MetaItemDelegate struct{}

func (d MetaItemDelegate) Height() int                             { return 1 }
func (d MetaItemDelegate) Spacing() int                            { return 0 }
func (d MetaItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d MetaItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	itemStyle := lipgloss.NewStyle().PaddingLeft(4)
	selecteditemStyle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("63"))
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
