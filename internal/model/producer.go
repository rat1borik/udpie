package model

import "github.com/google/uuid"

type Producer struct {
	Id         uuid.UUID  `json:"id"`
	UdpOptions UdpOptions `json:"udp_options"`
}

func NewProducer(udpOptions UdpOptions) *Producer {
	return &Producer{
		Id:         uuid.New(),
		UdpOptions: udpOptions,
	}
}
