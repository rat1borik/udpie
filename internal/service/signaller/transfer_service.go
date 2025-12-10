package signaller

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"udpie/internal/model"
	"udpie/internal/model/contract"
	"udpie/utils"
)

const defaultBlockSize = 1024 // 1KB default block size

type TransferService struct {
	mu              sync.RWMutex
	transfers       map[uuid.UUID]*model.Transfer
	fileService     contract.SignallerFileService
	producerService contract.SignallerProducerService
}

func NewTransferService(fileService contract.SignallerFileService, producerService contract.SignallerProducerService) *TransferService {
	return &TransferService{
		transfers:       make(map[uuid.UUID]*model.Transfer),
		fileService:     fileService,
		producerService: producerService,
	}
}

func (s *TransferService) InitTransfer(options contract.InitTransferOptions) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileMeta, err := s.fileService.GetFileMeta(options.FileId)
	if err != nil {
		return uuid.Nil, err
	}

	consumer := &model.Consumer{
		Id:         uuid.New(),
		UdpOptions: options.ConsumerUdpOptions,
	}

	transfer := model.NewTransfer(fileMeta, consumer, defaultBlockSize)
	s.transfers[transfer.Id] = transfer

	return transfer.Id, nil
}

func (s *TransferService) GetTransfer(id uuid.UUID) (*model.Transfer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transfer, exists := s.transfers[id]
	if !exists {
		return nil, errors.New("transfer not found")
	}

	return transfer, nil
}

func (s *TransferService) AcknowledgeBlocksProducer(transferId uuid.UUID, offset uint64, packets utils.BitArray) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transfer, exists := s.transfers[transferId]
	if !exists {
		return errors.New("transfer not found")
	}

	transfer.ReceivedBlocks.Or(offset, packets)
	return nil
}

func (s *TransferService) AcknowledgeBlocksConsumer(transferId uuid.UUID, offset uint64, packets utils.BitArray) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transfer, exists := s.transfers[transferId]
	if !exists {
		return errors.New("transfer not found")
	}

	transfer.ReceivedBlocks.Or(offset, packets)
	return nil
}
