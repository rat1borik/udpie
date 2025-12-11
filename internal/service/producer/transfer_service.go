package producer

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
	mu           sync.RWMutex
	transfers    map[uuid.UUID]*ActiveTransfer
	stateService *StateService
}

type ActiveTransfer struct {
	TransferId   uuid.UUID
	FileId       uuid.UUID
	FilePath     string
	BlockSize    uint64
	TotalBlocks  uint64
	ConsumerAddr *net.UDPAddr
	Status       string
	SentBlocks   map[uint64]bool
	mu           sync.Mutex
}

func NewTransferService(stateService *StateService) *TransferService {
	return &TransferService{
		transfers:    make(map[uuid.UUID]*ActiveTransfer),
		stateService: stateService,
	}
}

// StartTransfer starts sending a file to consumer
func (s *TransferService) StartTransfer(
	transferId uuid.UUID,
	fileId uuid.UUID,
	blockSize uint64,
	totalBlocks uint64,
	consumerAddr *net.UDPAddr,
) error {
	// Get file info from state
	fileInfo, exists := s.stateService.GetFile(fileId)
	if !exists {
		return fmt.Errorf("file not found in state: %s", fileId.String())
	}

	// Check if file exists
	if _, err := os.Stat(fileInfo.FilePath); err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	// Create active transfer
	transfer := &ActiveTransfer{
		TransferId:   transferId,
		FileId:       fileId,
		FilePath:     fileInfo.FilePath,
		BlockSize:    blockSize,
		TotalBlocks:  totalBlocks,
		ConsumerAddr: consumerAddr,
		Status:       "sending",
		SentBlocks:   make(map[uint64]bool),
	}

	s.mu.Lock()
	s.transfers[transferId] = transfer
	s.mu.Unlock()

	// Start sending in goroutine
	go s.sendFile(context.Background(), transfer)

	return nil
}

func (*TransferService) sendFile(ctx context.Context, transfer *ActiveTransfer) {
	file, err := os.Open(transfer.FilePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", transfer.FilePath, err)
		transfer.mu.Lock()
		transfer.Status = "failed"
		transfer.mu.Unlock()
		return
	}
	defer file.Close()

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, transfer.ConsumerAddr)
	if err != nil {
		fmt.Printf("Error creating UDP connection: %v\n", err)
		transfer.mu.Lock()
		transfer.Status = "failed"
		transfer.mu.Unlock()
		return
	}
	defer conn.Close()

	transferStartTime := time.Now()
	buffer := make([]byte, transfer.BlockSize)

	fmt.Printf("Starting file transfer: %s\n", transfer.FilePath)
	fmt.Printf("Transfer ID: %s\n", transfer.TransferId.String())
	fmt.Printf("Total blocks: %d\n", transfer.TotalBlocks)
	fmt.Printf("Block size: %d bytes\n", transfer.BlockSize)
	fmt.Printf("Consumer: %s\n", transfer.ConsumerAddr.String())

	// Send all blocks
	for blockNum := range uint64(transfer.TotalBlocks) {
		select {
		case <-ctx.Done():
			fmt.Printf("Transfer canceled: %s\n", transfer.TransferId.String())
			return
		default:
		}

		// Read block from file
		n, err := file.Read(buffer)
		if err != nil && n == 0 {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Error reading file at block %d: %v\n", blockNum, err)
			continue
		}

		// Create UDP packet
		packet := &client.UdpPacket{
			ContentType:  0x01, // Data packet
			SerialNumber: blockNum,
			TransferId:   transfer.TransferId,
			Timestamp:    time.Now(),
			Data:         buffer[:n],
		}

		// Marshal packet
		packetData, err := packet.Marshal(transferStartTime)
		if err != nil {
			fmt.Printf("Error marshaling packet %d: %v\n", blockNum, err)
			continue
		}

		// Send packet
		if _, err := conn.Write(packetData); err != nil {
			fmt.Printf("Error sending packet %d: %v\n", blockNum, err)
			continue
		}

		transfer.mu.Lock()
		transfer.SentBlocks[blockNum] = true
		transfer.mu.Unlock()

		if blockNum%100 == 0 || blockNum == transfer.TotalBlocks-1 {
			fmt.Printf("Sent block %d/%d\n", blockNum+1, transfer.TotalBlocks)
		}
	}

	transfer.mu.Lock()
	transfer.Status = "complete"
	transfer.mu.Unlock()

	fmt.Printf("File transfer completed: %s\n", transfer.TransferId.String())
}

// GetTransferStatus returns the status of a transfer
func (s *TransferService) GetTransferStatus(transferId uuid.UUID) (*ActiveTransfer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transfer, exists := s.transfers[transferId]
	return transfer, exists
}
