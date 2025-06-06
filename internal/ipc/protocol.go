// internal/ipc/protocol.go
package ipc

import (
	"encoding/json"
	"time"
)

// MessageType defines the type of IPC message
type MessageType string

const (
	MessageTypeLog        MessageType = "log"
	MessageTypeSourceInit MessageType = "source_init"
	MessageTypeSourceExit MessageType = "source_exit"
	MessageTypePing       MessageType = "ping"
	MessageTypePong       MessageType = "pong"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Level     LogLevel               `json:"level"`
	Content   string                 `json:"content"`
	Raw       string                 `json:"raw"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SourceInfo contains information about a log source
type SourceInfo struct {
	Name string `json:"name"`
	Type string `json:"type"` // "pipe", "docker", "podman"
}

// IPCMessage represents a message sent over the IPC channel
type IPCMessage struct {
	Type       MessageType `json:"type"`
	LogEntry   *LogEntry   `json:"log_entry,omitempty"`
	SourceInfo *SourceInfo `json:"source_info,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Marshal serializes an IPCMessage to JSON
func (m *IPCMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal deserializes JSON data into an IPCMessage
func (m *IPCMessage) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// NewLogMessage creates a new log message
func NewLogMessage(entry *LogEntry) *IPCMessage {
	return &IPCMessage{
		Type:     MessageTypeLog,
		LogEntry: entry,
	}
}

// NewSourceInitMessage creates a new source initialization message
func NewSourceInitMessage(name, sourceType string) *IPCMessage {
	return &IPCMessage{
		Type: MessageTypeSourceInit,
		SourceInfo: &SourceInfo{
			Name: name,
			Type: sourceType,
		},
	}
}

// NewSourceExitMessage creates a new source exit message
func NewSourceExitMessage(name string) *IPCMessage {
	return &IPCMessage{
		Type: MessageTypeSourceExit,
		SourceInfo: &SourceInfo{
			Name: name,
		},
	}
}
