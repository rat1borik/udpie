package model

import "github.com/google/uuid"

type Producer struct {
	Id         uuid.UUID  `json:"id"`
	UdpOptions UdpOptions `json:"udp_options"`
}
