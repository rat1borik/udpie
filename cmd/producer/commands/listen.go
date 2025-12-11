package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"

	"udpie/internal/config"
	"udpie/internal/service/producer"
)

// ListenCommand handles the listen command
type ListenCommand struct {
	cfg *config.ProducerConfig
}

func NewListenCommand(cfg *config.ProducerConfig) *ListenCommand {
	return &ListenCommand{cfg: cfg}
}

func (c *ListenCommand) Execute() {
	fs := flag.NewFlagSet("listen", flag.ExitOnError)
	signallerURL := fs.String("signaller", c.cfg.Signaller.URL, "Signaller server URL")
	producerIdStr := fs.String("producer-id", "", "Producer ID (optional, will use saved ID if not provided)")
	stateFile := fs.String("state-file", ".udpie-producer-state.json", "Path to state file")

	fs.Parse(os.Args[2:])

	// Initialize STUN service
	stunService := producer.NewSTUNService(c.cfg.STUN.Servers, c.cfg.STUN.LocalPort, c.cfg.STUN.Timeout)

	// Initialize state service
	stateService := producer.NewStateService(*stateFile)
	if err := stateService.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load state: %v\n", err)
	}

	// Get producer ID
	producerId, err := c.getProducerID(*producerIdStr, stateService)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create transfer service
	transferService := producer.NewTransferService(stateService)

	// Start websocket listener
	listener := producer.NewWebsocketListener(producerId, *signallerURL, stateService, transferService, stunService)
	if err := listener.Listen(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func (c *ListenCommand) getProducerID(producerIdStr string, stateService *producer.StateService) (uuid.UUID, error) {
	if producerIdStr != "" {
		producerId, err := uuid.Parse(producerIdStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid producer ID format: %w", err)
		}
		return producerId, nil
	}

	// Try to get from state
	savedId, exists := stateService.GetProducerId()
	if !exists {
		return uuid.Nil, fmt.Errorf("producer ID not found. Please provide -producer-id or register a producer first")
	}

	fmt.Printf("Using saved ProducerId: %s\n", savedId.String())
	return savedId, nil
}
