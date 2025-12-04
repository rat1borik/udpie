package model

import (
	"encoding/binary"
	"errors"
)

const serialNumberSize = 8

type UdpPacket struct {
	SerialNumber uint64
	Data         []byte
}

// Marshal serializes the UdpPacket into a byte slice.
// Format: [8 bytes: SerialNumber (big-endian)] [variable: Data]
func (u *UdpPacket) Marshal() []byte {
	buf := make([]byte, serialNumberSize+len(u.Data))
	binary.BigEndian.PutUint64(buf[:serialNumberSize], u.SerialNumber)
	copy(buf[serialNumberSize:], u.Data)
	return buf
}

// Unmarshal deserializes a byte slice into the UdpPacket.
// Returns an error if the data is too short to contain the serial number.
func (u *UdpPacket) Unmarshal(data []byte) error {
	if len(data) < serialNumberSize {
		return errors.New("data too short to contain serial number")
	}
	u.SerialNumber = binary.BigEndian.Uint64(data[:serialNumberSize])
	u.Data = make([]byte, len(data)-serialNumberSize)
	copy(u.Data, data[serialNumberSize:])
	return nil
}
