// internal/ui/keybindings.go
package ui

import "strings"

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Quit []string
	Help []string

	// Navigation
	NextPane     []string
	PrevPane     []string
	VimNav       []string
	DirectAccess []string

	// Layout
	CycleLayout []string
	Zoom        []string
	ZoomOut     []string

	// Search
	SearchLocal  []string
	SearchGlobal []string

	// Filter
	FilterError []string
	FilterWarn  []string
	FilterInfo  []string
	FilterAll   []string

	// Control
	Pause  []string
	Follow []string
	Clear  []string
	Export []string
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: []string{"q", "ctrl+c"},
		Help: []string{"?"},

		NextPane:     []string{"tab"},
		PrevPane:     []string{"shift+tab"},
		VimNav:       []string{"h", "j", "k", "l"},
		DirectAccess: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},

		CycleLayout: []string{"l"},
		Zoom:        []string{"z"},
		ZoomOut:     []string{"Z", "esc"},

		SearchLocal:  []string{"/"},
		SearchGlobal: []string{"ctrl+/", "?"},

		FilterError: []string{"e"},
		FilterWarn:  []string{"w"},
		FilterInfo:  []string{"i"},
		FilterAll:   []string{"a"},

		Pause:  []string{" "},
		Follow: []string{"f"},
		Clear:  []string{"c"},
		Export: []string{"x"},
	}
}

// Help returns a help text for the key bindings
func (k KeyMap) Help() string {
	help := []string{
		"Navigation:",
		"  1-9: Jump to pane",
		"  Tab/Shift+Tab: Cycle panes",
		"  h/j/k/l: Vim navigation",
		"",
		"Layout & View:",
		"  l: Cycle layouts",
		"  z: Zoom into pane",
		"  Z/Esc: Zoom out",
		"",
		"Search & Filter:",
		"  /: Search current pane",
		"  Ctrl+/: Global search",
		"  e/w/i/a: Filter by level",
		"",
		"Control:",
		"  Space: Pause/resume",
		"  f: Toggle follow mode",
		"  c: Clear current pane",
		"  q: Quit",
	}

	return strings.Join(help, "\n")
}
