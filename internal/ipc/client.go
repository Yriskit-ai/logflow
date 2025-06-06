// internal/ipc/client.go
package ipc

import (
	"encoding/json"
	"fmt"
	"net"
)

// Client handles IPC communication to the server
type Client struct {
	conn net.Conn
}

// NewClient creates a new IPC client
func NewClient() (*Client, error) {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to logflow daemon: %w", err)
	}

	return &Client{conn: conn}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendMessage sends an IPC message to the server
func (c *Client) SendMessage(msg *IPCMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	data = append(data, '\n') // Add newline for scanner
	_, err = c.conn.Write(data)
	return err
}

// InitSource initializes a source with the server
func (c *Client) InitSource(name, sourceType string) error {
	msg := NewSourceInitMessage(name, sourceType)
	return c.SendMessage(msg)
}

// SendLog sends a log entry to the server
func (c *Client) SendLog(entry *LogEntry) error {
	msg := NewLogMessage(entry)
	return c.SendMessage(msg)
}

// SendExit notifies the server that this source is exiting
func (c *Client) SendExit(sourceName string) error {
	msg := NewSourceExitMessage(sourceName)
	return c.SendMessage(msg)
}
