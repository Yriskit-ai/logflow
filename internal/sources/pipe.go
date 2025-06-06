// internal/sources/pipe.go
package sources

import (
	"bufio"
	"io"

	"github.com/yourusername/logflow/internal/ipc"
	"github.com/yourusername/logflow/internal/log"
)

// PipeSource reads logs from stdin/pipe
type PipeSource struct {
	name   string
	reader io.Reader
}

// NewPipeSource creates a new pipe source
func NewPipeSource(name string, reader io.Reader) *PipeSource {
	return &PipeSource{
		name:   name,
		reader: reader,
	}
}

// Name returns the source name
func (p *PipeSource) Name() string {
	return p.name
}

// Type returns the source type
func (p *PipeSource) Type() string {
	return "pipe"
}

// Stream reads from the pipe and sends log entries to the client
func (p *PipeSource) Stream(client *ipc.Client) error {
	scanner := bufio.NewScanner(p.reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Create log entry
		entry := log.NewLogEntry(p.name, line)

		// Convert to IPC format
		ipcEntry := &ipc.LogEntry{
			Timestamp: entry.Timestamp,
			Source:    entry.Source,
			Level:     ipc.LogLevel(entry.Level),
			Content:   entry.Content,
			Raw:       entry.Raw,
			Metadata:  entry.Metadata,
		}

		// Send to server
		if err := client.SendLog(ipcEntry); err != nil {
			return err
		}
	}

	return scanner.Err()
}
