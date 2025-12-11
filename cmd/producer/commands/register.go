package commands

import (
	"flag"
	"fmt"
	"net"
	"os"

	"udpie/internal/config"
	"udpie/internal/model"
	"udpie/internal/service/common"
	"udpie/internal/service/producer"
)

// RegisterCommand handles the register command
type RegisterCommand struct {
	cfg *config.ProducerConfig
}

func NewRegisterCommand(cfg *config.ProducerConfig) *RegisterCommand {
	return &RegisterCommand{cfg: cfg}
}

func (c *RegisterCommand) Execute() {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	stateFile := fs.String("state-file", ".udpie-producer-state.json", "Path to state file")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Initialize state service
	stateService := producer.NewStateService(*stateFile)
	if err := stateService.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load state: %v\n", err)
	}

	// Always use STUN to detect external address
	udpOptions, err := c.detectUDPOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Register producer using signaller URL from config
	producerService := producer.NewProducerService(c.cfg.Signaller.URL, stateService)
	producerId, err := producerService.Register(udpOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering producer: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Producer registered successfully\n")
	fmt.Printf("ProducerId: %s\n", producerId.String())
	fmt.Printf("State saved to: %s\n", *stateFile)
}

func (c *RegisterCommand) detectUDPOptions() (model.UdpOptions, error) {
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
