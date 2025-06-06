// internal/ui/layout.go
package ui

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

// renderHorizontalLayout renders panes stacked horizontally
func (a *App) renderHorizontalLayout(height int) string {
	if len(a.paneOrder) == 0 {
		return ""
	}

	var paneViews []string
	paneHeight := height / len(a.paneOrder)

	// Distribute remaining height to first few panes
	remainder := height % len(a.paneOrder)

	for i, paneName := range a.paneOrder {
		pane := a.panes[paneName]
		focused := (i == a.focusedPane)

		currentHeight := paneHeight
		if i < remainder {
			currentHeight++
		}

		paneView := pane.Render(a.width, currentHeight, focused, a.filterLevel, a.followMode)
		paneViews = append(paneViews, paneView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, paneViews...)
}

// renderVerticalLayout renders panes side by side vertically
func (a *App) renderVerticalLayout(height int) string {
	if len(a.paneOrder) == 0 {
		return ""
	}

	var paneViews []string
	paneWidth := a.width / len(a.paneOrder)

	// Distribute remaining width to first few panes
	remainder := a.width % len(a.paneOrder)

	for i, paneName := range a.paneOrder {
		pane := a.panes[paneName]
		focused := (i == a.focusedPane)

		currentWidth := paneWidth
		if i < remainder {
			currentWidth++
		}

		paneView := pane.Render(currentWidth, height, focused, a.filterLevel, a.followMode)
		paneViews = append(paneViews, paneView)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, paneViews...)
}

// renderGridLayout renders panes in a grid pattern
func (a *App) renderGridLayout(height int) string {
	if len(a.paneOrder) == 0 {
		return ""
	}

	// Calculate grid dimensions
	paneCount := len(a.paneOrder)
	cols := int(math.Ceil(math.Sqrt(float64(paneCount))))
	rows := int(math.Ceil(float64(paneCount) / float64(cols)))

	paneWidth := a.width / cols
	paneHeight := height / rows

	var gridRows []string

	for row := 0; row < rows; row++ {
		var rowPanes []string

		for col := 0; col < cols; col++ {
			paneIndex := row*cols + col
			if paneIndex >= len(a.paneOrder) {
				// Fill empty space
				emptyPane := lipgloss.NewStyle().
					Width(paneWidth).
					Height(paneHeight).
					Render("")
				rowPanes = append(rowPanes, emptyPane)
				continue
			}

			paneName := a.paneOrder[paneIndex]
			pane := a.panes[paneName]
			focused := (paneIndex == a.focusedPane)

			paneView := pane.Render(paneWidth, paneHeight, focused, a.filterLevel, a.followMode)
			rowPanes = append(rowPanes, paneView)
		}

		rowView := lipgloss.JoinHorizontal(lipgloss.Top, rowPanes...)
		gridRows = append(gridRows, rowView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, gridRows...)
}
