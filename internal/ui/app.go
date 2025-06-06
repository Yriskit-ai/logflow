// internal/ui/app.go
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/logflow/internal/ipc"
	"github.com/yourusername/logflow/internal/log"
)

// LayoutMode defines how panes are arranged
type LayoutMode int

const (
	LayoutHorizontal LayoutMode = iota
	LayoutVertical
	LayoutAutoGrid
)

// ViewMode defines the current view state
type ViewMode int

const (
	ViewMultiPane ViewMode = iota
	ViewZoomed
)

// SearchMode defines search scope
type SearchMode int

const (
	SearchNone   SearchMode = iota
	SearchLocal             // Search current pane only
	SearchGlobal            // Search all panes
)

// App represents the main TUI application
type App struct {
	server        *ipc.Server
	panes         map[string]*Pane
	paneOrder     []string
	layout        LayoutMode
	viewMode      ViewMode
	focusedPane   int
	zoomedPane    int
	searchMode    SearchMode
	searchQuery   string
	searchResults []SearchResult
	filterLevel   log.LogLevel
	followMode    bool
	paused        bool
	width         int
	height        int

	// Styles
	styles Styles
}

// SearchResult represents a search match across panes
type SearchResult struct {
	PaneName string
	Entry    log.LogEntry
	Index    int
}

// NewApp creates a new TUI application
func NewApp(server *ipc.Server) *App {
	return &App{
		server:      server,
		panes:       make(map[string]*Pane),
		paneOrder:   make([]string, 0),
		layout:      LayoutVertical,
		viewMode:    ViewMultiPane,
		focusedPane: 0,
		filterLevel: log.LogLevelDebug, // Show all levels by default
		followMode:  true,
		styles:      NewStyles(),
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	p := tea.NewProgram(a, tea.WithAltScreen())

	// Start listening for log entries
	go a.listenForLogs(p)

	return p.Run()
}

// Quit sends a quit message to the application
func (a *App) Quit() {
	// Implementation depends on how you want to handle shutdown
}

// listenForLogs processes incoming log entries from the IPC server
func (a *App) listenForLogs(p *tea.Program) {
	for entry := range a.server.LogChannel() {
		p.Send(LogEntryMsg{Entry: entry})
	}
}

// LogEntryMsg represents a new log entry message
type LogEntryMsg struct {
	Entry *ipc.LogEntry
}

// TickMsg for periodic updates
type TickMsg time.Time

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tick(),
	)
}

// tick returns a command that sends periodic tick messages
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateLayout()

	case tea.KeyMsg:
		return a.handleKeyPress(msg)

	case LogEntryMsg:
		a.handleLogEntry(msg.Entry)

	case TickMsg:
		cmds = append(cmds, tick())
	}

	return a, tea.Batch(cmds...)
}

// handleKeyPress processes keyboard input
func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "q" || msg.String() == "ctrl+c" {
		return a, tea.Quit
	}

	// Handle search mode
	if a.searchMode != SearchNone {
		return a.handleSearchInput(msg)
	}

	switch msg.String() {
	// Layout controls
	case "l":
		a.cycleLayout()

	// Zoom controls
	case "z":
		if a.viewMode == ViewMultiPane && len(a.paneOrder) > 0 {
			a.viewMode = ViewZoomed
			a.zoomedPane = a.focusedPane
		} else if a.viewMode == ViewZoomed {
			a.viewMode = ViewMultiPane
		}
		a.updateLayout()

	// Navigation
	case "tab":
		a.nextPane()
	case "shift+tab":
		a.prevPane()
	case "h":
		if a.layout == LayoutVertical {
			a.prevPane()
		}
	case "l":
		if a.layout == LayoutVertical {
			a.nextPane()
		}
	case "j":
		if a.layout == LayoutHorizontal {
			a.nextPane()
		} else {
			a.scrollDown()
		}
	case "k":
		if a.layout == LayoutHorizontal {
			a.prevPane()
		} else {
			a.scrollUp()
		}

	// Number keys for direct pane access
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		paneNum := int(msg.String()[0] - '1')
		if paneNum < len(a.paneOrder) {
			if a.viewMode == ViewZoomed {
				a.zoomedPane = paneNum
			}
			a.focusedPane = paneNum
			a.updateLayout()
		}

	// Search
	case "/":
		a.searchMode = SearchLocal
		a.searchQuery = ""
	case "ctrl+/", "?":
		a.searchMode = SearchGlobal
		a.searchQuery = ""

	// Filter controls
	case "e":
		a.filterLevel = log.LogLevelError
	case "w":
		a.filterLevel = log.LogLevelWarn
	case "i":
		a.filterLevel = log.LogLevelInfo
	case "a":
		a.filterLevel = log.LogLevelDebug

	// Control
	case " ":
		a.paused = !a.paused
	case "f":
		a.followMode = !a.followMode
	case "c":
		a.clearFocusedPane()
	}

	return a, nil
}

// handleLogEntry processes a new log entry
func (a *App) handleLogEntry(entry *ipc.LogEntry) {
	// Get or create pane for this source
	pane, exists := a.panes[entry.Source]
	if !exists {
		pane = NewPane(entry.Source, 1000) // Buffer size
		a.panes[entry.Source] = pane
		a.paneOrder = append(a.paneOrder, entry.Source)
		a.updateLayout()
	}

	// Convert IPC entry to internal log entry
	logEntry := log.LogEntry{
		Timestamp: entry.Timestamp,
		Source:    entry.Source,
		Level:     log.LogLevel(entry.Level),
		Content:   entry.Content,
		Raw:       entry.Raw,
		Metadata:  entry.Metadata,
	}

	// Add to pane if not paused
	if !a.paused {
		pane.AddEntry(logEntry)
	}
}

// View implements tea.Model
func (a *App) View() string {
	if len(a.paneOrder) == 0 {
		return a.styles.EmptyState.Render("Waiting for log sources...\n\nStart sending logs with:\npython app.py | logflow --source backend")
	}

	// Render header
	header := a.renderHeader()

	// Render main content based on view mode
	var content string
	if a.viewMode == ViewZoomed {
		content = a.renderZoomedView()
	} else {
		content = a.renderMultiPaneView()
	}

	// Render status bar
	status := a.renderStatusBar()

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		status,
	)
}

// renderHeader creates the application header
func (a *App) renderHeader() string {
	title := "logflow v1.0.0"
	sourceCount := fmt.Sprintf("%d sources", len(a.paneOrder))

	var layoutStr string
	switch a.layout {
	case LayoutHorizontal:
		layoutStr = "Horizontal"
	case LayoutVertical:
		layoutStr = "Vertical"
	case LayoutAutoGrid:
		layoutStr = "Grid"
	}

	if a.viewMode == ViewZoomed && len(a.paneOrder) > 0 {
		zoomedSource := a.paneOrder[a.zoomedPane]
		layoutStr = fmt.Sprintf("ZOOMED: [%d] %s", a.zoomedPane+1, zoomedSource)
	}

	controls := "[q]uit [l]ayout [z]oom [/]search [?]help"

	headerContent := fmt.Sprintf("%s │ %s │ %s │ %s", title, sourceCount, layoutStr, controls)

	return a.styles.Header.Width(a.width).Render(headerContent)
}

// renderMultiPaneView renders the multi-pane layout
func (a *App) renderMultiPaneView() string {
	if len(a.paneOrder) == 0 {
		return ""
	}

	contentHeight := a.height - 4 // Account for header and status bar

	switch a.layout {
	case LayoutHorizontal:
		return a.renderHorizontalLayout(contentHeight)
	case LayoutVertical:
		return a.renderVerticalLayout(contentHeight)
	case LayoutAutoGrid:
		return a.renderGridLayout(contentHeight)
	}

	return ""
}

// renderZoomedView renders a single pane in full screen
func (a *App) renderZoomedView() string {
	if a.zoomedPane >= len(a.paneOrder) {
		return "Invalid pane"
	}

	paneName := a.paneOrder[a.zoomedPane]
	pane := a.panes[paneName]

	contentHeight := a.height - 4
	return pane.Render(a.width, contentHeight, true, a.filterLevel, a.followMode)
}

// renderStatusBar creates the bottom status bar
func (a *App) renderStatusBar() string {
	var status []string

	// Active sources
	status = append(status, fmt.Sprintf("%d active sources", len(a.paneOrder)))

	// Filter level
	status = append(status, fmt.Sprintf("Filter: %s", a.filterLevel))

	// Search info
	if a.searchQuery != "" {
		if a.searchMode == SearchLocal {
			status = append(status, fmt.Sprintf("Search: /%s", a.searchQuery))
		} else {
			status = append(status, fmt.Sprintf("Global: /%s", a.searchQuery))
		}
	}

	// Current pane
	if len(a.paneOrder) > 0 {
		currentPane := a.paneOrder[a.focusedPane]
		status = append(status, fmt.Sprintf("Pane: %d (%s)", a.focusedPane+1, currentPane))
	}

	// Pause status
	if a.paused {
		status = append(status, "PAUSED")
	}

	statusText := strings.Join(status, " │ ")
	return a.styles.StatusBar.Width(a.width).Render(statusText)
}

// Layout helper methods
func (a *App) cycleLayout() {
	switch a.layout {
	case LayoutHorizontal:
		a.layout = LayoutVertical
	case LayoutVertical:
		a.layout = LayoutAutoGrid
	case LayoutAutoGrid:
		a.layout = LayoutHorizontal
	}
	a.updateLayout()
}

func (a *App) nextPane() {
	if len(a.paneOrder) > 0 {
		a.focusedPane = (a.focusedPane + 1) % len(a.paneOrder)
	}
}

func (a *App) prevPane() {
	if len(a.paneOrder) > 0 {
		a.focusedPane = (a.focusedPane - 1 + len(a.paneOrder)) % len(a.paneOrder)
	}
}

func (a *App) scrollDown() {
	if len(a.paneOrder) > 0 {
		paneName := a.paneOrder[a.focusedPane]
		if pane := a.panes[paneName]; pane != nil {
			pane.ScrollDown()
		}
	}
}

func (a *App) scrollUp() {
	if len(a.paneOrder) > 0 {
		paneName := a.paneOrder[a.focusedPane]
		if pane := a.panes[paneName]; pane != nil {
			pane.ScrollUp()
		}
	}
}

func (a *App) clearFocusedPane() {
	if len(a.paneOrder) > 0 {
		paneName := a.paneOrder[a.focusedPane]
		if pane := a.panes[paneName]; pane != nil {
			pane.Clear()
		}
	}
}

func (a *App) updateLayout() {
	// This would update pane dimensions based on current layout
	// Implementation depends on the specific layout algorithms
}

// Placeholder implementations for layout rendering
func (a *App) renderHorizontalLayout(height int) string {
	return "Horizontal layout implementation"
}

func (a *App) renderVerticalLayout(height int) string {
	return "Vertical layout implementation"
}

func (a *App) renderGridLayout(height int) string {
	return "Grid layout implementation"
}

func (a *App) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		a.performSearch()
		a.searchMode = SearchNone
	case "esc":
		a.searchMode = SearchNone
		a.searchQuery = ""
	case "backspace":
		if len(a.searchQuery) > 0 {
			a.searchQuery = a.searchQuery[:len(a.searchQuery)-1]
		}
	default:
		if len(msg.String()) == 1 {
			a.searchQuery += msg.String()
		}
	}
	return a, nil
}

func (a *App) performSearch() {
	// Implementation for search functionality
	a.searchResults = []SearchResult{}

	if a.searchMode == SearchLocal && len(a.paneOrder) > 0 {
		// Search current pane only
		paneName := a.paneOrder[a.focusedPane]
		if pane := a.panes[paneName]; pane != nil {
			// Implement search in pane
		}
	} else if a.searchMode == SearchGlobal {
		// Search all panes
		for _, paneName := range a.paneOrder {
			if pane := a.panes[paneName]; pane != nil {
				// Implement search across all panes
			}
		}
	}
}
