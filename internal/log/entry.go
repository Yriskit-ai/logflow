// internal/log/entry.go
package log

import (
	"fmt"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Level     LogLevel               `json:"level"`
	Content   string                 `json:"content"`
	Raw       string                 `json:"raw"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewLogEntry creates a new log entry from raw log line
func NewLogEntry(source, rawLine string) *LogEntry {
	parser := NewParser()

	entry := &LogEntry{
		Timestamp: time.Now(),
		Source:    source,
		Raw:       rawLine,
		Content:   rawLine,
		Metadata:  make(map[string]interface{}),
	}

	// Parse log level from the raw line
	entry.Level = parser.ParseLevel(rawLine)

	// Extract structured content if possible
	if structured := parser.ParseStructured(rawLine); structured != nil {
		if ts, ok := structured["timestamp"]; ok {
			if timestamp, ok := ts.(time.Time); ok {
				entry.Timestamp = timestamp
			}
		}
		if content, ok := structured["message"]; ok {
			if msg, ok := content.(string); ok {
				entry.Content = msg
			}
		}
		// Add other structured fields to metadata
		for k, v := range structured {
			if k != "timestamp" && k != "message" && k != "level" {
				entry.Metadata[k] = v
			}
		}
	}

	return entry
}

// String returns a formatted string representation of the log entry
func (e *LogEntry) String() string {
	timestamp := e.Timestamp.Format("15:04:05")
	return fmt.Sprintf("%s %s %s", timestamp, e.Level, e.Content)
}

// ColorString returns a colored string representation based on log level
func (e *LogEntry) ColorString() string {
	timestamp := e.Timestamp.Format("15:04:05")

	var levelColor string
	switch e.Level {
	case LogLevelError:
		levelColor = "\033[31m" // Red
	case LogLevelWarn:
		levelColor = "\033[33m" // Yellow
	case LogLevelInfo:
		levelColor = "\033[34m" // Blue
	case LogLevelDebug:
		levelColor = "\033[37m" // Gray
	default:
		levelColor = "\033[0m" // Reset
	}

	reset := "\033[0m"
	return fmt.Sprintf("%s %s%s%s %s", timestamp, levelColor, e.Level, reset, e.Content)
}
