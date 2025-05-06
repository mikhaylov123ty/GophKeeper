package utils

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (

	// Styles for different components and elements used in the application,
	// providing a consistent color scheme and formatting for the UI.
	CursorStyle     = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("205")).Bold(true) // Bright purple
	SelectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)            // Green
	UnselectedStyle = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("7"))              // White
	BackgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("245"))                       // Grey background
	TitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))             // Bold yellow
	SeparatorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))                       // Light grey

	ListHeight = 15

	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"

	ColorRed = "\033[31m"
)

// NavigateFooter returns a styled string with navigation instructions for the user to interact with the menu options.
func NavigateFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nUse arrow keys to navigate and enter to select.\n")) // Navigation instructions
}

// AddItemsFooter returns a string containing the footer instructions for adding items, styled with background and separator styles.
func AddItemsFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress Enter to save, CTRL+Q to cancel, or Backspace to delete the last character.\n"))
}

// ListItemsFooter returns a styled string displaying navigation and action instructions for a list interface.
func ListItemsFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nUse arrow keys to navigate. E to edit. D to delete. Enter to select. CTRL+Q to cancel.\n"))
}

// ItemDataFooter renders a styled footer with a prompt to return using CTRL+Q.
func ItemDataFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress CTRL+Q to return.\n"))
}

// BinaryItemDataFooter renders a styled footer with instructions for downloading a file or returning to the previous screen.
func BinaryItemDataFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress D to download file to output folder. CTRL+Q to return.\n"))
}

// AuthFooter returns a styled footer string providing instructions for navigating and exiting the authentication screen.
func AuthFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress Tab to switch fields, Enter to submit, or CTRL+Q to exit.\n"))
}

func DataHeader() string {
	separator := "\n" + strings.Repeat("=", 40) + "\n" // Creates a separator line for better readability
	body := TitleStyle.Render("Information", separator)

	return body
}
