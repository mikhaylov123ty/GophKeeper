package screens

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"strings"
)

type ErrorScreen struct {
	backScreen models.Screen
	err        error
}

func (screen *ErrorScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		default:
			return screen.backScreen, nil
		}
	}
	return screen.backScreen, nil
}

func (screen *ErrorScreen) View() string {
	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sError%s\n"+
			"=======================%s"+
			"%s%s%s\n",
		utils.ColorBold, utils.ColorReset,
		separator,
		utils.ColorGreen, screen.err.Error(), utils.ColorReset)
}
