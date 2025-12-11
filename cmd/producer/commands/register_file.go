package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"udpie/internal/config"
	"udpie/internal/service/producer"
)

// RegisterFileCommand handles the register-file command
type RegisterFileCommand struct {
	cfg *config.ProducerConfig
}

func NewRegisterFileCommand(cfg *config.ProducerConfig) *RegisterFileCommand {
	return &RegisterFileCommand{cfg: cfg}
}

func (c *RegisterFileCommand) Execute() {
	fs := flag.NewFlagSet("register-file", flag.ExitOnError)
	signallerURL := fs.String("signaller", c.cfg.Signaller.URL, "Signaller server URL")
	producerIdStr := fs.String("producer-id", "", "Producer ID (optional, will use saved ID if not provided)")
	filePath := fs.String("path", "", "Path to file (required)")
	stateFile := fs.String("state-file", ".udpie-producer-state.json", "Path to state file")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *filePath == "" {
		fmt.Fprintf(os.Stderr, "Error: -path is required\n")
		fs.Usage()
		os.Exit(1)
	}

	// Validate and get file info
	absPath, fileName, fileSize, err := c.validateFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize state service
	stateService := producer.NewStateService(*stateFile)
	if loadErr := stateService.Load(); loadErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load state: %v\n", loadErr)
	}

	// Get producer ID
	producerId, err := c.getProducerID(*producerIdStr, stateService)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Register file
	producerService := producer.NewProducerService(*signallerURL, stateService)
	fileId, err := producerService.RegisterFile(fileName, fileSize, producerId, absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File registered successfully\n")
	fmt.Printf("FileId: %s\n", fileId.String())
	fmt.Printf("FilePath: %s\n", absPath)
	fmt.Printf("State saved to: %s\n", *stateFile)
}

func (*RegisterFileCommand) validateFile(filePath string) (absPath, fileName string, fileSize uint64, err error) {
	// Get absolute path
	absPath, err = filepath.Abs(filePath)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", 0, fmt.Errorf("file does not exist: %s", absPath)
		}
		return "", "", 0, fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return "", "", 0, fmt.Errorf("path is a directory, not a file: %s", absPath)
	}

	fileName = fileInfo.Name()
	// nolint:gosec // fileInfo.Size() returns int64, safe to convert to uint64 for file sizes
	fileSize = uint64(fileInfo.Size())

	return absPath, fileName, fileSize, nil
}

func (*RegisterFileCommand) getProducerID(producerIdStr string, stateService *producer.StateService) (uuid.UUID, error) {
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
