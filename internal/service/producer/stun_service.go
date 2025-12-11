package producer

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v2"
)

type STUNService struct {
	servers  []string
	localPort int
	timeout   time.Duration
}

func NewSTUNService(servers []string, localPort int, timeoutSeconds int) *STUNService {
	return &STUNService{
		servers:   servers,
		localPort: localPort,
		timeout:   time.Duration(timeoutSeconds) * time.Second,
	}
}

// Query queries STUN servers to determine the external IP and port
func (s *STUNService) Query() (net.Addr, error) {
	// Try each STUN server until one succeeds
	for _, serverAddr := range s.servers {
		addr, err := s.queryServer(serverAddr)
		if err != nil {
			continue
		}
		return addr, nil
	}

	return nil, fmt.Errorf("all STUN servers failed")
}

func (s *STUNService) queryServer(serverAddr string) (net.Addr, error) {
	// Resolve STUN server address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve STUN server: %w", err)
	}

	// Create local UDP connection
	localAddr := &net.UDPAddr{Port: s.localPort}
	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial STUN server: %w", err)
	}
	defer conn.Close()

	// Build STUN binding request
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// Send STUN request
	if _, err := conn.Write(message.Raw); err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read STUN response
	var buf [1500]byte
	n, _, err := conn.ReadFrom(buf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %w", err)
	}

	// Decode STUN message
	m := &stun.Message{Raw: buf[:n]}
	if err := m.Decode(); err != nil {
		return nil, fmt.Errorf("failed to decode STUN message: %w", err)
	}

	// Extract XOR-Mapped-Address
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(m); err != nil {
		return nil, fmt.Errorf("failed to get XOR-Mapped-Address: %w", err)
	}

	return &net.UDPAddr{
		IP:   xorAddr.IP,
		Port: xorAddr.Port,
	}, nil
}

