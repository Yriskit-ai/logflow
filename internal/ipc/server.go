// internal/ipc/server.go
package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

const SocketPath = "/tmp/logflow.sock"

// Server handles IPC communication from source processes
type Server struct {
	listener net.Listener
	clients  map[net.Conn]*Client
	mutex    sync.RWMutex
	logChan  chan *LogEntry
	quit     chan struct{}
}

// NewServer creates a new IPC server
func NewServer() (*Server, error) {
	// Remove existing socket file
	os.Remove(SocketPath)

	// Create socket directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(SocketPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create unix socket: %w", err)
	}

	server := &Server{
		listener: listener,
		clients:  make(map[net.Conn]*Client),
		logChan:  make(chan *LogEntry, 1000), // Buffered channel
		quit:     make(chan struct{}),
	}

	go server.acceptConnections()
	return server, nil
}

// LogChannel returns the channel for receiving log entries
func (s *Server) LogChannel() <-chan *LogEntry {
	return s.logChan
}

// Close shuts down the server
func (s *Server) Close() error {
	close(s.quit)

	s.mutex.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.mutex.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}

	os.Remove(SocketPath)
	return nil
}

// acceptConnections handles incoming client connections
func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}
			go s.handleClient(conn)
		}
	}
}

// handleClient processes messages from a connected client
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	client := &Client{conn: conn}

	s.mutex.Lock()
	s.clients[conn] = client
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.clients, conn)
		s.mutex.Unlock()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg IPCMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		switch msg.Type {
		case MessageTypeLog:
			if msg.LogEntry != nil {
				select {
				case s.logChan <- msg.LogEntry:
				default:
					// Channel full, drop message
				}
			}
		case MessageTypeSourceInit:
			// Handle source initialization
		case MessageTypeSourceExit:
			// Handle source exit
		}
	}
}
