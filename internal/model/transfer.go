package model

import (
	"math"

	"github.com/google/uuid"

	"udpie/utils"
)

type TransferStatus string

const (
	TransferStatusCreated           TransferStatus = "created"
	TransferStatusProducerAccepting TransferStatus = "producer_accepting"
	TransferStatusFailed            TransferStatus = "failed"
	TransferStatusHolePunching      TransferStatus = "hole_punching"
	TransferStatusDataSending       TransferStatus = "data_sending"
	TransferStatusComplete          TransferStatus = "complete"
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
}

func NewTransfer(meta *FileMeta, consumer *Consumer, blockSize uint64) *Transfer {
	blocks := uint64(math.Ceil(float64(meta.Size) / float64(blockSize)))
	return &Transfer{
		Id:             uuid.New(),
		FileMeta:       meta,
		Consumer:       consumer,
		Status:         TransferStatusCreated,
		TotalBlocks:    blocks,
		BlockSize:      blockSize,
		FailedBlocks:   utils.NewBitArray(blocks),
		ReceivedBlocks: utils.NewBitArray(blocks),
	}
}
