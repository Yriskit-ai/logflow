// internal/sources/docker.go
package sources

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/Yriskit-ai/logflow/internal/ipc"
	"github.com/Yriskit-ai/logflow/internal/log"
)

// DockerSource reads logs from a Docker container
type DockerSource struct {
	name        string
	containerID string
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewDockerSource creates a new Docker source
func NewDockerSource(name, containerID string) *DockerSource {
	ctx, cancel := context.WithCancel(context.Background())

	return &DockerSource{
		name:        name,
		containerID: containerID,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Name returns the source name
func (d *DockerSource) Name() string {
	return d.name
}

// Type returns the source type
func (d *DockerSource) Type() string {
	return "docker"
}

// Stream starts following Docker container logs
func (d *DockerSource) Stream(client *ipc.Client) error {
	// Start docker logs command
	d.cmd = exec.CommandContext(d.ctx, "docker", "logs", "-f", "--timestamps", d.containerID)

	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := d.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker logs command: %w", err)
	}

	// Stream stdout
	go d.streamPipe(client, stdout, "stdout")

	// Stream stderr
	go d.streamPipe(client, stderr, "stderr")

	// Wait for command to finish
	return d.cmd.Wait()
}

// streamPipe handles streaming from a pipe
func (d *DockerSource) streamPipe(client *ipc.Client, pipe io.Reader, stream string) {
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse Docker timestamp format: 2023-01-01T12:00:00.000000000Z message
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
		entry := log.NewLogEntry(d.name, content)
		entry.Timestamp = timestamp

		// Add stream metadata
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata["stream"] = stream
		entry.Metadata["container_id"] = d.containerID

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

// Close stops the Docker logs command
func (d *DockerSource) Close() error {
	if d.cancel != nil {
		d.cancel()
	}
	if d.cmd != nil && d.cmd.Process != nil {
		return d.cmd.Process.Kill()
	}
	return nil
}
