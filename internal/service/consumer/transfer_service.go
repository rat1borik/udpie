package consumer

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"udpie/internal/client"
)

type TransferService struct {
	mu        sync.RWMutex
	transfers map[uuid.UUID]*ActiveTransfer
}

type ActiveTransfer struct {
	TransferId    uuid.UUID
	FilePath      string
	BlockSize     uint64
	TotalBlocks   uint64
	ProducerAddr  *net.UDPAddr
	Status        string
	ReceivedBlocks map[uint64][]byte
	mu            sync.Mutex
}

func NewTransferService() *TransferService {
	return &TransferService{
		transfers: make(map[uuid.UUID]*ActiveTransfer),
	}
}

// StartTransfer starts receiving a file from producer
func (s *TransferService) StartTransfer(
	transferId uuid.UUID,
	filePath string,
	blockSize uint64,
	totalBlocks uint64,
	producerAddr *net.UDPAddr,
) error {
	// Create active transfer
	transfer := &ActiveTransfer{
		TransferId:    transferId,
		FilePath:      filePath,
		BlockSize:     blockSize,
		TotalBlocks:   totalBlocks,
		ProducerAddr:  producerAddr,
		Status:        "receiving",
		ReceivedBlocks: make(map[uint64][]byte),
	}

	s.mu.Lock()
	s.transfers[transferId] = transfer
	s.mu.Unlock()

	// Start receiving in goroutine
	go s.receiveFile(context.Background(), transfer)

	return nil
}

func (s *TransferService) receiveFile(ctx context.Context, transfer *ActiveTransfer) {
	// Create UDP connection to listen for packets
	localAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Printf("Error creating UDP listener: %v\n", err)
		transfer.mu.Lock()
		transfer.Status = "failed"
		transfer.mu.Unlock()
		return
	}
	defer conn.Close()

	transferStartTime := time.Now()
	buffer := make([]byte, 65507) // Max UDP packet size

	fmt.Printf("Starting file download: %s\n", transfer.FilePath)
	fmt.Printf("Transfer ID: %s\n", transfer.TransferId.String())
	fmt.Printf("Total blocks: %d\n", transfer.TotalBlocks)
	fmt.Printf("Block size: %d bytes\n", transfer.BlockSize)
	fmt.Printf("Producer: %s\n", transfer.ProducerAddr.String())
	fmt.Printf("Listening on: %s\n", conn.LocalAddr().String())

	// Receive packets
	receivedCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Transfer cancelled: %s\n", transfer.TransferId.String())
			return
		default:
		}

		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			fmt.Printf("Error setting read deadline: %v\n", err)
			continue
		}

		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// Check if we've received all blocks
			transfer.mu.Lock()
			if len(transfer.ReceivedBlocks) >= int(transfer.TotalBlocks) {
				transfer.mu.Unlock()
				break
			}
			transfer.mu.Unlock()

			// Check if timeout (expected when no more packets)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Check if we should continue waiting
				transfer.mu.Lock()
				received := len(transfer.ReceivedBlocks)
				transfer.mu.Unlock()

				if received >= int(transfer.TotalBlocks) {
					break
				}
				// Continue waiting for more packets
				continue
			}

			fmt.Printf("Error reading UDP packet: %v\n", err)
			continue
		}

		// Verify sender address matches producer
		if addr.IP.String() != transfer.ProducerAddr.IP.String() || addr.Port != transfer.ProducerAddr.Port {
			// Not from expected producer, skip
			continue
		}

		// Parse UDP packet
		packet := &client.UdpPacket{}
		if err := packet.Unmarshal(buffer[:n], transferStartTime); err != nil {
			fmt.Printf("Error unmarshaling packet: %v\n", err)
			continue
		}

		// Verify transfer ID matches
		if packet.TransferId != transfer.TransferId {
			continue
		}

		// Store received block
		transfer.mu.Lock()
		transfer.ReceivedBlocks[packet.SerialNumber] = packet.Data
		receivedCount++
		transfer.mu.Unlock()

		if receivedCount%100 == 0 || receivedCount == int(transfer.TotalBlocks) {
			fmt.Printf("Received block %d/%d\n", receivedCount, transfer.TotalBlocks)
		}

		// Check if we've received all blocks
		if receivedCount >= int(transfer.TotalBlocks) {
			break
		}
	}

	// Write file
	if err := s.writeFile(transfer); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		transfer.mu.Lock()
		transfer.Status = "failed"
		transfer.mu.Unlock()
		return
	}

	transfer.mu.Lock()
	transfer.Status = "complete"
	transfer.mu.Unlock()

	fmt.Printf("File download completed: %s\n", transfer.TransferId.String())
	fmt.Printf("File saved to: %s\n", transfer.FilePath)
}

func (s *TransferService) writeFile(transfer *ActiveTransfer) error {
	// Create output file
	file, err := os.Create(transfer.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write blocks in order
	for blockNum := uint64(0); blockNum < transfer.TotalBlocks; blockNum++ {
		blockData, exists := transfer.ReceivedBlocks[blockNum]
		if !exists {
			// Missing block, fill with zeros
			fmt.Printf("Warning: Missing block %d, filling with zeros\n", blockNum)
			blockData = make([]byte, transfer.BlockSize)
		}

		if _, err := file.Write(blockData); err != nil {
			return fmt.Errorf("failed to write block %d: %w", blockNum, err)
		}
	}

	return nil
}

// GetTransferStatus returns the status of a transfer
func (s *TransferService) GetTransferStatus(transferId uuid.UUID) (*ActiveTransfer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transfer, exists := s.transfers[transferId]
	return transfer, exists
}

// GetStatus returns the status string of a transfer
func (t *ActiveTransfer) GetStatus() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status
}

