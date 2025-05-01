package utils

import "github.com/charmbracelet/lipgloss"

var (
	CursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))                     // Bright purple
	SelectedStyle   = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("2")).Bold(true) // Green
	UnselectedStyle = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("7"))            // White
	BackgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("245"))                     // Grey background
	TitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("65"))           // Bold yellow
	SeparatorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))                     // Light grey

	ListHeight = 15

	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
)

func NavigateFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nUse arrow keys to navigate and enter to select.\n")) // Navigation instructions
}

func AddItemsFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress Enter to save, CTRL+Q to cancel, or Backspace to delete the last character.\n"))
}

func ListItemsFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nUse arrow keys to navigate. E to edit. D to delete. Enter to select. CTRL+Q to cancel.\n"))
}

func ItemDataFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress CTRL+Q to return.\n"))
}

func BinaryItemDataFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress D to download file to output folder. CTRL+Q to return.\n"))
}

func AuthFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress Tab to switch fields, Enter to submit, or CTRL+Q to exit.\n"))
}
