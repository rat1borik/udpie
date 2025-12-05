package client

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

const (
	contentTypeSize  = 1
	serialNumberSize = 4
	transferIdSize   = 16
	timestampSize    = 4
	dataSizeSize     = 2
	headerSize       = contentTypeSize + serialNumberSize + transferIdSize + timestampSize + dataSizeSize
)

type UdpPacket struct {
	ContentType  byte
	SerialNumber uint64
	TransferId   uuid.UUID
	Timestamp    time.Time
	Data         []byte
}

// Marshal serializes the UdpPacket into a byte slice.
// Format: [1 byte: ContentType] [4 bytes: SerialNumber (big-endian)] [16 bytes: TransferId]
// [4 bytes: Timestamp (relative to transferStartTime, big-endian)] [2 bytes: DataSize (big-endian)] [variable: Data]
func (u *UdpPacket) Marshal(transferStartTime time.Time) ([]byte, error) {
	lenData := len(u.Data)
	if lenData > math.MaxUint16 {
		return nil, errors.New("data size too large")
	}

	buf := make([]byte, headerSize+len(u.Data))
	offset := 0

	buf[offset] = u.ContentType
	offset += contentTypeSize

	if u.SerialNumber > math.MaxUint32 {
		return nil, errors.New("data size too large")
	}

	binary.BigEndian.PutUint32(buf[offset:offset+serialNumberSize], uint32(u.SerialNumber))
	offset += serialNumberSize

	copy(buf[offset:offset+transferIdSize], u.TransferId[:])
	offset += transferIdSize

	//nolint:gosec
	timestampMs := uint64(u.Timestamp.UnixMilli() - transferStartTime.UnixMilli())
	if timestampMs > math.MaxUint32 {
		return nil, errors.New("timestamp too large")
	}

	binary.BigEndian.PutUint32(buf[offset:offset+timestampSize], uint32(timestampMs))
	offset += timestampSize

	//nolint:gosec
	binary.BigEndian.PutUint16(buf[offset:offset+dataSizeSize], uint16(lenData))
	offset += dataSizeSize

	copy(buf[offset:], u.Data)

	return buf, nil
}

// Unmarshal deserializes a byte slice into the UdpPacket.
// Returns an error if the data is too short to contain the header.
func (u *UdpPacket) Unmarshal(data []byte, transferStartTime time.Time) error {
	if len(data) < headerSize {
		return errors.New("data too short to contain packet header")
	}

	offset := 0

	u.ContentType = data[offset]
	offset += contentTypeSize

	u.SerialNumber = uint64(binary.BigEndian.Uint32(data[offset : offset+serialNumberSize]))
	offset += serialNumberSize

	copy(u.TransferId[:], data[offset:offset+transferIdSize])
	offset += transferIdSize

	millis := binary.BigEndian.Uint32(data[offset : offset+timestampSize])
	timestampMs := transferStartTime.UnixMilli() + int64(millis)
	u.Timestamp = time.UnixMilli(timestampMs)

	offset += timestampSize

	//nolint:gosec
	lenData := binary.BigEndian.Uint16(data[offset : offset+dataSizeSize])
	offset += dataSizeSize

	//nolint:gosec
	if len(data) != headerSize+int(lenData) {
		return errors.New("declared data size not equals available data size")
	}

	//nolint:gosec
	u.Data = make([]byte, lenData)

	//nolint:gosec
	copy(u.Data, data[offset:offset+int(lenData)])

	return nil
}
