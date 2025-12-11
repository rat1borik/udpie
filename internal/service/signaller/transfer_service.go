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
	mu               sync.RWMutex
	transfers        map[uuid.UUID]*model.Transfer
	fileService      contract.SignallerFileService
	producerService  contract.SignallerProducerService
	websocketService contract.WebsocketProducerService
}

func NewTransferService(fileService contract.SignallerFileService,
	producerService contract.SignallerProducerService,
	websocketService contract.WebsocketProducerService) *TransferService {
	return &TransferService{
		transfers:        make(map[uuid.UUID]*model.Transfer),
		fileService:      fileService,
		producerService:  producerService,
		websocketService: websocketService,
	}
}

func (s *TransferService) InitTransfer(options contract.InitTransferOptions) (*contract.InitTransferResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileMeta, err := s.fileService.GetFileMeta(options.FileId)
	if err != nil {
		return nil, err
	}

	consumer := &model.Consumer{
		Id:         uuid.New(),
		UdpOptions: options.ConsumerUdpOptions,
	}

	transfer := model.NewTransfer(fileMeta, consumer, defaultBlockSize)
	s.transfers[transfer.Id] = transfer
	transfer.Status = model.TransferStatusCreated

	resp, err := s.websocketService.MakeClientRequestWithTimeout(&contract.WebsocketRequest{
		ProducerId: fileMeta.ProducerId,
		Type:       "init_transfer",
		Data: model.ProducerInitTransferRequestData{
			FileId:             fileMeta.Id,
			BlockSize:          transfer.FileMeta.Size,
			BlocksCount:        transfer.TotalBlocks,
			ConsumerId:         consumer.Id,
			ConsumerUdpOptions: consumer.UdpOptions,
		},
	}, contract.DefaultWebsocketRequestTimeout)
	if err != nil {
		transfer.Status = model.TransferStatusFailed
		return nil, err
	}

	respData, ok := resp.Data.(model.ProducerInitTransferResponseData)
	if !ok {
		transfer.Status = model.TransferStatusFailed
		return nil, errors.New("invalid response data")
	}

	if respData.Status == model.RequestTransferStatusRejected {
		transfer.Status = model.TransferStatusProducerRejected
		return nil, errors.New("producer rejected transfer")
	}

	transfer.Status = model.TransferStatusProducerAccepted
	if err := s.producerService.UpdateUdpOptions(fileMeta.ProducerId, respData.ProducerUdpOptions); err != nil {
		return nil, err
	}

	return &contract.InitTransferResult{
		TransferId:         transfer.Id,
		ProducerUdpOptions: respData.ProducerUdpOptions,
		BlockSize:          transfer.BlockSize,
		TotalBlocks:        transfer.TotalBlocks,
	}, nil
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
