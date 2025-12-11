package commands

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"

	"udpie/internal/config"
	"udpie/internal/model"
	"udpie/internal/service/common"
	"udpie/internal/service/consumer"
)

// DownloadCommand handles the download command
type DownloadCommand struct {
	cfg *config.ProducerConfig
}

func NewDownloadCommand(cfg *config.ProducerConfig) *DownloadCommand {
	return &DownloadCommand{cfg: cfg}
}

// nolint:gocyclo,funlen // complex download logic with multiple error handling paths
func (c *DownloadCommand) Execute() {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	fileIdStr := fs.String("file-id", "", "File ID to download (required)")
	outputPath := fs.String("output", "", "Output file path (optional, defaults to file name)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *fileIdStr == "" {
		fmt.Fprintf(os.Stderr, "Error: -file-id is required\n")
		fs.Usage()
		os.Exit(1)
	}

	fileId, err := uuid.Parse(*fileIdStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid file ID format: %v\n", err)
		os.Exit(1)
	}

	// Always use STUN to detect external address
	udpOptions, err := c.detectUDPOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize consumer service using signaller URL from config
	consumerService := consumer.NewConsumerService(c.cfg.Signaller.URL)

	// Initiate download
	fmt.Printf("Initiating download for file ID: %s\n", fileId.String())
	transferResult, err := consumerService.InitDownload(fileId, udpOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initiating download: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download initiated successfully\n")
	fmt.Printf("Transfer ID: %s\n", transferResult.TransferId.String())
	fmt.Printf("Producer: %s:%d\n", transferResult.ProducerUdpOptions.ExternalIp, transferResult.ProducerUdpOptions.ExternalPort)
	fmt.Printf("Block Size: %d bytes\n", transferResult.BlockSize)
	fmt.Printf("Total Blocks: %d\n", transferResult.TotalBlocks)

	// Determine output file path
	outputFilePath := *outputPath
	if outputFilePath == "" {
		// Use file ID as default name
		outputFilePath = fileId.String()
	}

	// Make path absolute
	absPath, err := filepath.Abs(outputFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	// Create directory if needed
	dir := filepath.Dir(absPath)
	const dirPerm = 0755
	if mkdirErr := os.MkdirAll(dir, dirPerm); mkdirErr != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create directory: %v\n", mkdirErr)
		os.Exit(1)
	}

	// Resolve producer address
	producerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d",
		transferResult.ProducerUdpOptions.ExternalIp,
		transferResult.ProducerUdpOptions.ExternalPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving producer address: %v\n", err)
		os.Exit(1)
	}

	// Start receiving file
	transferService := consumer.NewTransferService()
	if err := transferService.StartTransfer(
		transferResult.TransferId,
		absPath,
		transferResult.BlockSize,
		transferResult.TotalBlocks,
		producerAddr,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting transfer: %v\n", err)
		os.Exit(1)
	}

	// Setup interrupt handler
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	// Wait for transfer to complete
	fmt.Println("\nWaiting for file transfer to complete...")
	fmt.Println("Press Ctrl+C to cancel")

	// Wait for transfer to complete or interrupt
	done := make(chan bool, 1)
	go func() {
		for {
			transfer, exists := transferService.GetTransferStatus(transferResult.TransferId)
			if !exists {
				const checkInterval = 100 * time.Millisecond
				time.Sleep(checkInterval)
				continue
			}
			status := transfer.GetStatus()
			if status == "complete" || status == "failed" {
				done <- true
				return
			}
			const waitInterval = 500 * time.Millisecond
			time.Sleep(waitInterval)
		}
	}()

	// Wait for completion or interrupt
	select {
	case <-done:
		transfer, exists := transferService.GetTransferStatus(transferResult.TransferId)
		if exists {
			status := transfer.GetStatus()
			if status == "complete" {
				fmt.Println("\nFile download completed successfully!")
				os.Exit(0)
			} else {
				fmt.Fprintf(os.Stderr, "\nFile download failed!\n")
				os.Exit(1)
			}
		}
	case <-interruptChan:
		fmt.Println("\nDownload canceled by user")
		os.Exit(1)
	}
}

func (c *DownloadCommand) detectUDPOptions() (model.UdpOptions, error) {
	// Always use STUN to detect external IP and port
	fmt.Println("Detecting external IP and port via STUN...")

	stunService := common.NewSTUNService(c.cfg.STUN.Servers, c.cfg.STUN.LocalPort, c.cfg.STUN.Timeout)
	extAddr, err := stunService.Query()
	if err != nil {
		return model.UdpOptions{}, fmt.Errorf("failed to detect external address via STUN: %w", err)
	}

	udpAddr := extAddr.(*net.UDPAddr)
	fmt.Printf("Detected external address: %s:%d\n", udpAddr.IP.String(), udpAddr.Port)

	return model.UdpOptions{
		ExternalIp:   udpAddr.IP.String(),
		ExternalPort: udpAddr.Port,
	}, nil
}
