package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// TODO make unexport fields
type MetaItem struct {
	Id          uuid.UUID
	Title       string
	Description string
	dataID      string
	Created     string // You can use time.Time for actual timestamp
	Modified    string // Same as above
}

func (m MetaItem) FilterValue() string { return "" }

type metaItemDelegate struct{}

func (d metaItemDelegate) Height() int                             { return 1 }
func (d metaItemDelegate) Spacing() int                            { return 0 }
func (d metaItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d metaItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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
