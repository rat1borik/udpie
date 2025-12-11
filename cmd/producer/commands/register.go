package commands

import (
	"flag"
	"fmt"
	"net"
	"os"

	"udpie/internal/config"
	"udpie/internal/model"
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
	signallerURL := fs.String("signaller", c.cfg.Signaller.URL, "Signaller server URL")
	externalIP := fs.String("ip", "", "External IP address (auto-detected via STUN if not provided)")
	externalPort := fs.Int("port", 0, "External port (auto-detected via STUN if not provided)")
	localPort := fs.Int("local-port", c.cfg.STUN.LocalPort, "Local UDP port for STUN query")
	useSTUN := fs.Bool("stun", true, "Use STUN to auto-detect external IP and port")
	stateFile := fs.String("state-file", ".udpie-producer-state.json", "Path to state file")

	fs.Parse(os.Args[2:])

	// Initialize state service
	stateService := producer.NewStateService(*stateFile)
	if err := stateService.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load state: %v\n", err)
	}

	// Detect external address if needed
	udpOptions, err := c.detectUDPOptions(*externalIP, *externalPort, *localPort, *useSTUN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fs.Usage()
		os.Exit(1)
	}

	// Register producer
	producerService := producer.NewProducerService(*signallerURL, stateService)
	producerId, err := producerService.Register(udpOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering producer: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Producer registered successfully\n")
	fmt.Printf("ProducerId: %s\n", producerId.String())
	fmt.Printf("State saved to: %s\n", *stateFile)
}

func (c *RegisterCommand) detectUDPOptions(externalIP string, externalPort int, localPort int, useSTUN bool) (model.UdpOptions, error) {
	// If IP and port are not provided, try to detect via STUN
	if (externalIP == "" || externalPort == 0) && useSTUN {
		fmt.Println("Detecting external IP and port via STUN...")

		stunService := producer.NewSTUNService(c.cfg.STUN.Servers, localPort, c.cfg.STUN.Timeout)
		extAddr, err := stunService.Query()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting external address via STUN: %v\n", err)
			if externalIP == "" || externalPort == 0 {
				return model.UdpOptions{}, fmt.Errorf("ip and port are required when STUN fails")
			}
		} else {
			udpAddr := extAddr.(*net.UDPAddr)
			if externalIP == "" {
				externalIP = udpAddr.IP.String()
			}
			if externalPort == 0 {
				externalPort = udpAddr.Port
			}
			fmt.Printf("Detected external address: %s:%d\n", externalIP, externalPort)
		}
	}

	if externalIP == "" || externalPort == 0 {
		return model.UdpOptions{}, fmt.Errorf("ip and port are required")
	}

	return model.UdpOptions{
		ExternalIp:   externalIP,
		ExternalPort: externalPort,
	}, nil
}
