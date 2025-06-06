package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Yriskit-ai/logflow/internal/ipc"
	"github.com/Yriskit-ai/logflow/internal/sources"
	"github.com/Yriskit-ai/logflow/internal/ui"
	"github.com/spf13/cobra"
)

var (
	sourceName      string
	dockerContainer string
	podmanContainer string
)

var rootCmd = &cobra.Command{
	Use:   "logflow",
	Short: "Multi-source log viewer for development",
	Long: `logflow is a TUI application that aggregates and displays logs from multiple sources.
	
Examples:
  logflow                                    # Start the dashboard
  python app.py | logflow --source backend  # Pipe logs to dashboard
  logflow --docker redis --source redis     # Attach to Docker container`,
	Run: runDashboard,
}

func init() {
	rootCmd.Flags().StringVarP(&sourceName, "source", "s", "", "Source name for this log stream")
	rootCmd.Flags().StringVar(&dockerContainer, "docker", "", "Docker container name/ID to attach to")
	rootCmd.Flags().StringVar(&podmanContainer, "podman", "", "Podman container name/ID to attach to")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDashboard(cmd *cobra.Command, args []string) {
	// If source name is provided, we're a feeder process
	if sourceName != "" {
		runSourceFeeder()
		return
	}

	// If container flags are provided, attach to container
	if dockerContainer != "" {
		runContainerFeeder("docker", dockerContainer)
		return
	}

	if podmanContainer != "" {
		runContainerFeeder("podman", podmanContainer)
		return
	}

	// Otherwise, start the main TUI dashboard
	startTUIDashboard()
}

func runSourceFeeder() {
	client, err := ipc.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to logflow daemon: %v", err)
	}
	defer client.Close()

	// Initialize the source
	if err := client.InitSource(sourceName, "pipe"); err != nil {
		log.Fatalf("Failed to initialize source: %v", err)
	}

	// Create pipe source and start feeding
	pipeSource := sources.NewPipeSource(sourceName, os.Stdin)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		client.Close()
		os.Exit(0)
	}()

	// Start streaming logs
	if err := pipeSource.Stream(client); err != nil {
		log.Fatalf("Failed to stream logs: %v", err)
	}
}

func runContainerFeeder(containerType, containerID string) {
	client, err := ipc.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to logflow daemon: %v", err)
	}
	defer client.Close()

	// Initialize the source
	if err := client.InitSource(sourceName, containerType); err != nil {
		log.Fatalf("Failed to initialize source: %v", err)
	}

	// Create container source based on type
	var containerSource sources.Source
	switch containerType {
	case "docker":
		containerSource = sources.NewDockerSource(sourceName, containerID)
	case "podman":
		containerSource = sources.NewPodmanSource(sourceName, containerID)
	default:
		log.Fatalf("Unknown container type: %s", containerType)
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		client.Close()
		os.Exit(0)
	}()

	// Start streaming logs
	if err := containerSource.Stream(client); err != nil {
		log.Fatalf("Failed to stream container logs: %v", err)
	}
}

func startTUIDashboard() {
	// Start the IPC server
	server, err := ipc.NewServer()
	if err != nil {
		log.Fatalf("Failed to start IPC server: %v", err)
	}

	// Start the TUI application
	app := ui.NewApp(server)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		app.Quit()
		server.Close()
		os.Exit(0)
	}()

	// Run the TUI
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run TUI: %v", err)
	}
}
