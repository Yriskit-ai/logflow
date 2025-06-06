// internal/ui/styles.go
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles contains all the styling for the application
type Styles struct {
	Header     lipgloss.Style
	StatusBar  lipgloss.Style
	EmptyState lipgloss.Style

	// Pane styles
	PaneFocused   lipgloss.Style
	PaneUnfocused lipgloss.Style
	PaneHeader    lipgloss.Style

	// Log level styles
	ErrorLevel lipgloss.Style
	WarnLevel  lipgloss.Style
	InfoLevel  lipgloss.Style
	DebugLevel lipgloss.Style

	// Search styles
	SearchHighlight lipgloss.Style
	SearchQuery     lipgloss.Style
}

// NewStyles creates a new styles instance
func NewStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true),

		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1),

		EmptyState: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(2).
			Align(lipgloss.Center),

		PaneFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1),

		PaneUnfocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),

		PaneHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1),

		ErrorLevel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true),

		WarnLevel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")),

		InfoLevel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")),

		DebugLevel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),

		SearchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("11")).
			Foreground(lipgloss.Color("0")),

		SearchQuery: lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1),
	}
}
