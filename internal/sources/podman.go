// internal/sources/podman.go
package sources

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/yourusername/logflow/internal/ipc"
	"github.com/yourusername/logflow/internal/log"
)

// PodmanSource reads logs from a Podman container
type PodmanSource struct {
	name        string
	containerID string
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewPodmanSource creates a new Podman source
func NewPodmanSource(name, containerID string) *PodmanSource {
	ctx, cancel := context.WithCancel(context.Background())

	return &PodmanSource{
		name:        name,
		containerID: containerID,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Name returns the source name
func (p *PodmanSource) Name() string {
	return p.name
}

// Type returns the source type
func (p *PodmanSource) Type() string {
	return "podman"
}

// Stream starts following Podman container logs
func (p *PodmanSource) Stream(client *ipc.Client) error {
	// Start podman logs command
	p.cmd = exec.CommandContext(p.ctx, "podman", "logs", "-f", "--timestamps", p.containerID)

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start podman logs command: %w", err)
	}

	// Stream stdout
	go p.streamPipe(client, stdout, "stdout")

	// Stream stderr
	go p.streamPipe(client, stderr, "stderr")

	// Wait for command to finish
	return p.cmd.Wait()
}

// streamPipe handles streaming from a pipe
func (p *PodmanSource) streamPipe(client *ipc.Client, pipe io.Reader, stream string) {
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse Podman timestamp format (similar to Docker)
		var timestamp time.Time
		var content string

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			if ts, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
				timestamp = ts
				content = parts[1]
			} else {
				timestamp = time.Now()
				content = line
			}
		} else {
			timestamp = time.Now()
			content = line
		}

		// Create log entry
		entry := log.NewLogEntry(p.name, content)
		entry.Timestamp = timestamp

		// Add stream metadata
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata["stream"] = stream
		entry.Metadata["container_id"] = p.containerID

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
		client.SendLog(ipcEntry)
	}
}

// Close stops the Podman logs command
func (p *PodmanSource) Close() error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}
