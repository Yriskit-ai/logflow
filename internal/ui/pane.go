// internal/ui/pane.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/logflow/internal/log"
)

// Pane represents a single log display pane
type Pane struct {
	name       string
	buffer     *log.Buffer
	scrollPos  int
	width      int
	height     int
	focused    bool
	lastSearch string
}

// NewPane creates a new log pane
func NewPane(name string, bufferSize int) *Pane {
	return &Pane{
		name:   name,
		buffer: log.NewBuffer(bufferSize),
	}
}

// AddEntry adds a log entry to the pane
func (p *Pane) AddEntry(entry log.LogEntry) {
	p.buffer.Add(entry)
}

// Render renders the pane content
func (p *Pane) Render(width, height int, focused bool, filterLevel log.LogLevel, followMode bool) string {
	p.width = width
	p.height = height
	p.focused = focused

	// Get filtered entries
	entries := p.buffer.Filter(filterLevel)

	// Calculate visible area
	contentHeight := height - 2 // Account for borders
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Auto-scroll to bottom if follow mode is enabled
	if followMode && len(entries) > contentHeight {
		p.scrollPos = len(entries) - contentHeight
	}

	// Ensure scroll position is valid
	if p.scrollPos < 0 {
		p.scrollPos = 0
	}
	if len(entries) > 0 && p.scrollPos > len(entries)-contentHeight {
		p.scrollPos = len(entries) - contentHeight
		if p.scrollPos < 0 {
			p.scrollPos = 0
		}
	}

	// Get visible entries
	var visibleEntries []log.LogEntry
	if len(entries) > 0 {
		start := p.scrollPos
		end := start + contentHeight
		if end > len(entries) {
			end = len(entries)
		}
		if start < end {
			visibleEntries = entries[start:end]
		}
	}

	// Render entries
	var lines []string
	for _, entry := range visibleEntries {
		line := p.formatLogEntry(entry, width-4) // Account for borders and padding
		lines = append(lines, line)
	}

	// Fill remaining space with empty lines
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	// Create pane header
	header := p.renderHeader()

	// Apply styling based on focus state
	var style lipgloss.Style
	if focused {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")). // Bright blue
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")). // Gray
			Padding(0, 1)
	}

	// Combine header and content
	paneContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
	)

	return style.Width(width).Height(height).Render(paneContent)
}

// renderHeader creates the pane header with name and stats
func (p *Pane) renderHeader() string {
	count := p.buffer.Count()
	status := "●●●" // Active indicator

	if count >= 1000 {
		countStr := fmt.Sprintf("%.1fk lines", float64(count)/1000)
		return fmt.Sprintf("%s %s - %s", status, p.name, countStr)
	}

	return fmt.Sprintf("%s %s - %d lines", status, p.name, count)
}

// formatLogEntry formats a log entry for display
func (p *Pane) formatLogEntry(entry log.LogEntry, maxWidth int) string {
	timestamp := entry.Timestamp.Format("15:04:05")

	// Get level color
	var levelStyle lipgloss.Style
	switch entry.Level {
	case log.LogLevelError:
		levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
	case log.LogLevelWarn:
		levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	case log.LogLevelInfo:
		levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	case log.LogLevelDebug:
		levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	default:
		levelStyle = lipgloss.NewStyle()
	}

	// Format the line
	levelStr := levelStyle.Render(string(entry.Level))
	line := fmt.Sprintf("%s %s %s", timestamp, levelStr, entry.Content)

	// Truncate if too long
	if len(line) > maxWidth {
		line = line[:maxWidth-3] + "..."
	}

	return line
}

// ScrollDown scrolls the pane down
func (p *Pane) ScrollDown() {
	entries := p.buffer.GetAll()
	maxScroll := len(entries) - (p.height - 2)
	if maxScroll < 0 {
		maxScroll = 0
	}

	if p.scrollPos < maxScroll {
		p.scrollPos++
	}
}

// ScrollUp scrolls the pane up
func (p *Pane) ScrollUp() {
	if p.scrollPos > 0 {
		p.scrollPos--
	}
}

// Clear clears all entries in the pane
func (p *Pane) Clear() {
	p.buffer.Clear()
	p.scrollPos = 0
}

// Search searches for a term in the pane
func (p *Pane) Search(term string) []log.LogEntry {
	p.lastSearch = term
	return p.buffer.Search(term)
}

// GetEntryCount returns the number of entries in the pane
func (p *Pane) GetEntryCount() int {
	return p.buffer.Count()
}
