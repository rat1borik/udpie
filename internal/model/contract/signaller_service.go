package contract

import (
	"github.com/google/uuid"

	"udpie/internal/model"
	"udpie/utils"
)

type RegisterFileOptions struct {
	Name  string          `json:"name"`
	Size  uint64          `json:"size"`
	Owner *model.Producer `json:"owner"`
}

type RegisterProducerOptions struct {
	UdpOptions model.UdpOptions `json:"udp_options"`
}

type InitTransferOptions struct {
	FileId             uuid.UUID        `json:"file_id"`
	ConsumerUdpOptions model.UdpOptions `json:"consumer_udp_options"`
}

type SignallerProducerService interface {
	RegisterProducer(options RegisterProducerOptions) (uuid.UUID, error)
	GetProducer(id uuid.UUID) (*model.Producer, error)
}

type SignallerFileService interface {
	RegisterFile(options RegisterFileOptions) (uuid.UUID, error)
	GetFileMeta(id uuid.UUID) (*model.FileMeta, error)
}

type SignallerTransferService interface {
	InitTransfer(options InitTransferOptions) (uuid.UUID, error)
	GetTransfer(id uuid.UUID) (*model.Transfer, error)
	AcknowledgeBlocksProducer(transferId uuid.UUID, offset uint64, packets utils.BitArray) error
	AcknowledgeBlocksConsumer(transferId uuid.UUID, offset uint64, packets utils.BitArray) error
}
