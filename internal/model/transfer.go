package model

import (
	"math"

	"github.com/google/uuid"

	"udpie/utils"
)

type TransferStatus string

const (
	TransferStatusCreated          TransferStatus = "created"
	TransferStatusFailed           TransferStatus = "failed"
	TransferStatusProducerAccepted TransferStatus = "producer_accepted"
	TransferStatusProducerRejected TransferStatus = "producer_rejected"
	TransferStatusDataSending      TransferStatus = "data_sending"
	TransferStatusComplete         TransferStatus = "complete"
)

type RequestTransferStatus string

const (
	RequestTransferStatusAccepted RequestTransferStatus = "accepted"
	RequestTransferStatusRejected RequestTransferStatus = "rejected"
)

type BlockStatus string

type Transfer struct {
	Id             uuid.UUID      `json:"id"`
	FileMeta       *FileMeta      `json:"file_meta"`
	Consumer       *Consumer      `json:"consumer"`
	Status         TransferStatus `json:"status"`
	TotalBlocks    uint64         `json:"total_blocks"`
	BlockSize      uint64         `json:"block_size"`
	FailedBlocks   utils.BitArray `json:"failed_blocks"`
	ReceivedBlocks utils.BitArray `json:"received_blocks"`
	SentBlocks     utils.BitArray `json:"sent_blocks"`
}

type ProducerInitTransferRequestData struct {
	FileId             uuid.UUID  `json:"file_id"`
	BlockSize          uint64     `json:"block_size"`
	BlocksCount        uint64     `json:"blocks_count"`
	ConsumerId         uuid.UUID  `json:"consumer_id"`
	ConsumerUdpOptions UdpOptions `json:"consumer_udp_options"`
}

type ProducerInitTransferResponseData struct {
	Status             RequestTransferStatus `json:"status"`
	ProducerUdpOptions UdpOptions            `json:"producer_udp_options"`
}

func NewTransfer(meta *FileMeta, consumer *Consumer, blockSize uint64) *Transfer {
	blocks := uint64(math.Ceil(float64(meta.Size) / float64(blockSize)))
	return &Transfer{
		Id:          uuid.New(),
		FileMeta:    meta,
		Consumer:    consumer,
		Status:      TransferStatusCreated,
		TotalBlocks: blocks,
		BlockSize:   blockSize,
		// FailedBlocks:   utils.NewBitArray(blocks),
		// ReceivedBlocks: utils.NewBitArray(blocks),
		// SentBlocks:     utils.NewBitArray(blocks),
	}
}
