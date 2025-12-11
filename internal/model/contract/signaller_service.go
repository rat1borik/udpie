package contract

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"

	"udpie/internal/model"
)

type RegisterFileOptions struct {
	Name       string    `json:"name"`
	Size       uint64    `json:"size"`
	Hash       []byte    `json:"hash"`
	ProducerId uuid.UUID `json:"producer_id"`
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
	UpdateUdpOptions(id uuid.UUID, options model.UdpOptions) error
}

type SignallerFileService interface {
	RegisterFile(options RegisterFileOptions) (uuid.UUID, error)
	GetFileMeta(id uuid.UUID) (*model.FileMeta, error)
}

type InitTransferResult struct {
	TransferId         uuid.UUID        `json:"transfer_id"`
	ProducerUdpOptions model.UdpOptions `json:"producer_udp_options"`
	BlockSize          uint64           `json:"block_size"`
	TotalBlocks        uint64           `json:"total_blocks"`
}

type SignallerTransferService interface {
	InitTransfer(options InitTransferOptions) (*InitTransferResult, error)
	GetTransfer(id uuid.UUID) (*model.Transfer, error)
}

// WebsocketRequest represents a websocket request message
type WebsocketRequest struct {
	ProducerId uuid.UUID `json:"producer_id"`
	RequestID  string    `json:"request_id,omitempty"` // Auto-generated if empty
	Type       string    `json:"type"`
	Data       any       `json:"data,omitempty"`
}

// WebsocketResponse represents a successful websocket response
type WebsocketResponse struct {
	ProducerId uuid.UUID `json:"producer_id"`
	RequestID  string    `json:"request_id"`
	Data       any       `json:"data,omitempty"`
}

const (
	DefaultWebsocketRequestTimeout = 3 * time.Second
)

type WebsocketProducerService interface {
	HandleConnection(producerId uuid.UUID, conn *websocket.Conn) error
	MakeClientRequestWithTimeout(request *WebsocketRequest, timeout time.Duration) (*WebsocketResponse, error)
}
