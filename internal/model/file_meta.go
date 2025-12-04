package model

import "github.com/google/uuid"

type FileMeta struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Size  uint64    `json:"size"`
	Hash  []byte    `json:"hash"`
	Owner *Producer `json:"owner"`
}
