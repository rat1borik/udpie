package producer

import (
	"fmt"

	"github.com/google/uuid"

	"udpie/internal/client"
	"udpie/internal/model"
)

type ProducerService struct {
	signallerClient *client.SignallerClient
	stateService    *StateService
}

func NewProducerService(signallerURL string, stateService *StateService) *ProducerService {
	return &ProducerService{
		signallerClient: client.NewSignallerClient(signallerURL),
		stateService:    stateService,
	}
}

// Register registers a producer with the given UDP options and saves the ID
func (s *ProducerService) Register(udpOptions model.UdpOptions) (uuid.UUID, error) {
	producerId, err := s.signallerClient.RegisterProducer(udpOptions)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to register producer: %w", err)
	}

	// Save producer ID to state
	if err := s.stateService.SetProducerId(producerId); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save producer ID: %w", err)
	}

	return producerId, nil
}

// RegisterFile registers a file for a producer and saves the file path
func (s *ProducerService) RegisterFile(name string, size uint64, producerId uuid.UUID, filePath string) (uuid.UUID, error) {
	fileId, err := s.signallerClient.RegisterFile(name, size, producerId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to register file: %w", err)
	}

	// Save file info to state
	if err := s.stateService.AddFile(fileId, name, size, filePath); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save file info: %w", err)
	}

	return fileId, nil
}
