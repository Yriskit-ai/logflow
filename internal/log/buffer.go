// internal/log/buffer.go
package log

import (
	"strings"
	"sync"
)

// Buffer manages a circular buffer of log entries for a source
type Buffer struct {
	entries []LogEntry
	size    int
	index   int
	count   int
	mutex   sync.RWMutex
}

// NewBuffer creates a new log buffer with the specified size
func NewBuffer(size int) *Buffer {
	return &Buffer{
		entries: make([]LogEntry, size),
		size:    size,
		index:   0,
		count:   0,
	}
}

// Add appends a log entry to the buffer
func (b *Buffer) Add(entry LogEntry) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.entries[b.index] = entry
	b.index = (b.index + 1) % b.size

	if b.count < b.size {
		b.count++
	}
}

// GetAll returns all log entries in chronological order
func (b *Buffer) GetAll() []LogEntry {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if b.count == 0 {
		return nil
	}

	result := make([]LogEntry, b.count)

	if b.count < b.size {
		// Buffer not full yet, entries are from 0 to count-1
		copy(result, b.entries[:b.count])
	} else {
		// Buffer is full, entries wrap around
		startIdx := b.index
		copy(result, b.entries[startIdx:])
		copy(result[b.size-startIdx:], b.entries[:startIdx])
	}

	return result
}

// GetRecent returns the most recent n log entries
func (b *Buffer) GetRecent(n int) []LogEntry {
	all := b.GetAll()
	if len(all) <= n {
		return all
	}
	return all[len(all)-n:]
}

// Clear removes all entries from the buffer
func (b *Buffer) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.count = 0
	b.index = 0
}

// Count returns the number of entries in the buffer
func (b *Buffer) Count() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.count
}

// Filter returns entries matching the specified log level or higher
func (b *Buffer) Filter(minLevel LogLevel) []LogEntry {
	all := b.GetAll()
	var filtered []LogEntry

	levelOrder := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	minLevelOrder := levelOrder[minLevel]

	for _, entry := range all {
		if levelOrder[entry.Level] >= minLevelOrder {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// Search returns entries containing the specified search term
func (b *Buffer) Search(term string) []LogEntry {
	all := b.GetAll()
	var matches []LogEntry

	for _, entry := range all {
		if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(term)) ||
			strings.Contains(strings.ToLower(entry.Raw), strings.ToLower(term)) {
			matches = append(matches, entry)
		}
	}

	return matches
}
