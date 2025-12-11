package model

import (
	"github.com/google/uuid"
)

type FileMeta struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Size       uint64    `json:"size"`
	Hash       []byte    `json:"hash"`
	ProducerId uuid.UUID `json:"producer_id"`
}

func NewFileMeta(name string, size uint64, hash []byte, producerId uuid.UUID) *FileMeta {
	return &FileMeta{
		Id:         uuid.New(),
		Name:       name,
		Size:       size,
		Hash:       hash,
		ProducerId: producerId,
	}
}

func (f *FileMeta) Clone() *FileMeta {
	return &FileMeta{
		Id:         f.Id,
		Name:       f.Name,
		Size:       f.Size,
		Hash:       f.Hash,
		ProducerId: f.ProducerId,
	}
}
