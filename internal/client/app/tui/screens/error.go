package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
)

const (
	errorLength = 60
)

// ErrorScreen represents a screen for displaying error messages and navigating back to a previous screen.
type ErrorScreen struct {
	backScreen models.Screen
	err        error
}

// Update processes the received message, handles user input, and returns the updated screen or a command to execute.
func (screen *ErrorScreen) Update(msg tea.Msg) (models.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "\x00" {
		}
	}
	return screen.backScreen, nil
}

// View renders the error message on the ErrorScreen with line breaks and formatting for better readability.
func (screen *ErrorScreen) View() string {
	errStrings := strings.Split(screen.err.Error(), " ")
	var errJoin string
	length := 0
	for i := range errStrings {
		errJoin = strings.Join([]string{errJoin, errStrings[i]}, " ")
		length += len(errStrings[i])
		if length > errorLength {
			length = 0
			errJoin = fmt.Sprintf("%s\n", errJoin)
		}
	}

	separator := "\n" + strings.Repeat("-", 40) + "\n" // Creates a separator line for better readability
	return fmt.Sprintf(
		"%sError%s"+
			"%s"+
			"%s%s%s\n\n"+
			"%s",
		utils.ColorRed, utils.ColorReset,
		separator,
		utils.ColorGreen, errJoin, utils.ColorReset,
		utils.ItemDataFooter(),
	)
}
