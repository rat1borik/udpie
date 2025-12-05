package client

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUdpPacket_Marshal(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)
	testTransferId := uuid.New()

	tests := []struct {
		name    string
		packet  *UdpPacket
		wantErr bool
		errMsg  string
	}{
		{
			name: "basic packet with data",
			packet: &UdpPacket{
				ContentType:  0x01,
				SerialNumber: 12345,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000), // 1000ms after transfer start
				Data:         []byte("hello world"),
			},
			wantErr: false,
		},
		{
			name: "packet with empty data",
			packet: &UdpPacket{
				ContentType:  0x02,
				SerialNumber: 0,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte{},
			},
			wantErr: false,
		},
		{
			name: "packet with large data",
			packet: &UdpPacket{
				ContentType:  0x03,
				SerialNumber: 999,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         make([]byte, 1000),
			},
			wantErr: false,
		},
		{
			name: "packet with max uint32 serial number",
			packet: &UdpPacket{
				ContentType:  0x04,
				SerialNumber: math.MaxUint32,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte("test"),
			},
			wantErr: false,
		},
		{
			name: "packet with max uint32 timestamp offset",
			packet: &UdpPacket{
				ContentType:  0x05,
				SerialNumber: 100,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000000000 + int64(math.MaxUint32)),
				Data:         []byte("test"),
			},
			wantErr: false,
		},
		{
			name: "packet with serial number exceeding uint32",
			packet: &UdpPacket{
				ContentType:  0x06,
				SerialNumber: math.MaxUint32 + 1,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte("test"),
			},
			wantErr: true,
			errMsg:  "data size too large",
		},
		{
			name: "packet with timestamp offset exceeding uint32",
			packet: &UdpPacket{
				ContentType:  0x07,
				SerialNumber: 100,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000000000 + int64(math.MaxUint32) + 1),
				Data:         []byte("test"),
			},
			wantErr: true,
			errMsg:  "timestamp too large",
		},
		{
			name: "packet with zero timestamp offset",
			packet: &UdpPacket{
				ContentType:  0x08,
				SerialNumber: 42,
				TransferId:   testTransferId,
				Timestamp:    transferStartTime, // Same as transfer start time
				Data:         []byte("zero offset"),
			},
			wantErr: false,
		},
		{
			name: "packet with various content types",
			packet: &UdpPacket{
				ContentType:  0xFF,
				SerialNumber: 255,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte("max content type"),
			},
			wantErr: false,
		},
		{
			name: "packet with different transfer ID",
			packet: &UdpPacket{
				ContentType:  0x09,
				SerialNumber: 123,
				TransferId:   uuid.New(),
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte("different transfer"),
			},
			wantErr: false,
		},
		{
			name: "packet with data size exceeding uint16",
			packet: &UdpPacket{
				ContentType:  0x0A,
				SerialNumber: 100,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         make([]byte, math.MaxUint16+1), // Exceeds uint16 max
			},
			wantErr: true,
			errMsg:  "data size too large",
		},
		{
			name: "packet with max uint16 data size",
			packet: &UdpPacket{
				ContentType:  0x0B,
				SerialNumber: 100,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         make([]byte, math.MaxUint16), // Max uint16
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.packet.Marshal(transferStartTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("Marshal() error = %v, wantErr %v", err.Error(), tt.errMsg)
				}
				return
			}

			// Verify marshaled data size
			expectedSize := headerSize + len(tt.packet.Data)
			if len(data) != expectedSize {
				t.Errorf("Marshal() data size = %d, want %d", len(data), expectedSize)
			}

			// Verify DataSize was set correctly
			if tt.packet.DataSize != uint64(len(tt.packet.Data)) {
				t.Errorf("Marshal() DataSize = %d, want %d", tt.packet.DataSize, len(tt.packet.Data))
			}
		})
	}
}

func TestUdpPacket_Unmarshal(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "data too short for header",
			data:    []byte{0x01, 0x02, 0x03}, // Only 3 bytes, need at least headerSize
			wantErr: true,
			errMsg:  "data too short to contain packet header",
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "data too short to contain packet header",
		},
		{
			name:    "header only, no data",
			data:    nil,   // Will be set in test
			wantErr: false, // Should succeed with empty data
		},
		{
			name:    "header with declared data size larger than available",
			data:    nil, // Will be set in test
			wantErr: true,
			errMsg:  "data too short: declared data size exceeds available data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := &UdpPacket{}
			var err error
			testData := tt.data

			// For the "header only" case, we need to set DataSize to 0 in the header
			if tt.name == "header only, no data" {
				// Create proper header with DataSize = 0
				testData = make([]byte, headerSize)
				testData[0] = 0x01 // ContentType
				// SerialNumber = 0 (already zero)
				// TransferId = 0 (already zero)
				// Timestamp offset = 0 (already zero)
				// DataSize = 0 (already zero)
			}

			// For the "header with declared data size larger than available" case
			if tt.name == "header with declared data size larger than available" {
				// Create buffer with only 5 bytes of data, but set DataSize to 100
				testData = make([]byte, headerSize+5)
				testData[0] = 0x01 // ContentType
				// SerialNumber = 0 (already zero)
				// TransferId = 0 (already zero)
				// Timestamp offset = 0 (already zero)
				// Set DataSize to 100 (larger than available 5 bytes)
				dataSizeOffset := contentTypeSize + serialNumberSize + transferIdSize + timestampSize
				binary.BigEndian.PutUint16(testData[dataSizeOffset:dataSizeOffset+dataSizeSize], 100)
			}

			err = packet.Unmarshal(testData, transferStartTime)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("Unmarshal() error = %v, wantErr %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestUdpPacket_Marshal_Unmarshal_RoundTrip(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)
	testTransferId := uuid.New()

	tests := []struct {
		name   string
		packet *UdpPacket
	}{
		{
			name: "basic round trip",
			packet: &UdpPacket{
				ContentType:  0x01,
				SerialNumber: 12345,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000), // 1000ms after transfer start
				Data:         []byte("hello world"),
			},
		},
		{
			name: "empty data round trip",
			packet: &UdpPacket{
				ContentType:  0x02,
				SerialNumber: 0,
				TransferId:   testTransferId,
				Timestamp:    transferStartTime, // Same as transfer start
				Data:         []byte{},
			},
		},
		{
			name: "large data round trip",
			packet: &UdpPacket{
				ContentType:  0x03,
				SerialNumber: 999999,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         make([]byte, 1000),
			},
		},
		{
			name: "max uint32 serial number round trip",
			packet: &UdpPacket{
				ContentType:  math.MaxUint8,
				SerialNumber: math.MaxUint32,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000000000 + int64(math.MaxUint32)),
				Data:         []byte("max values"),
			},
		},
		{
			name: "various content types",
			packet: &UdpPacket{
				ContentType:  0xAB,
				SerialNumber: 42,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte("content type test"),
			},
		},
		{
			name: "binary data",
			packet: &UdpPacket{
				ContentType:  0x04,
				SerialNumber: 100,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte{0x00, 0xFF, 0x42, 0x7F, 0x80, 0x01},
			},
		},
		{
			name: "single byte data",
			packet: &UdpPacket{
				ContentType:  0x05,
				SerialNumber: 1,
				TransferId:   testTransferId,
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte{0x42},
			},
		},
		{
			name: "different transfer IDs",
			packet: &UdpPacket{
				ContentType:  0x06,
				SerialNumber: 2024,
				TransferId:   uuid.New(),
				Timestamp:    time.UnixMilli(1000001000),
				Data:         []byte("different transfer id"),
			},
		},
		{
			name: "zero timestamp offset",
			packet: &UdpPacket{
				ContentType:  0x07,
				SerialNumber: 100,
				TransferId:   testTransferId,
				Timestamp:    transferStartTime, // Same as transfer start
				Data:         []byte("zero offset"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := tt.packet.Marshal(transferStartTime)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal
			unmarshaled := &UdpPacket{}
			err = unmarshaled.Unmarshal(data, transferStartTime)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare fields
			if unmarshaled.ContentType != tt.packet.ContentType {
				t.Errorf("ContentType = %v, want %v", unmarshaled.ContentType, tt.packet.ContentType)
			}
			if unmarshaled.SerialNumber != tt.packet.SerialNumber {
				t.Errorf("SerialNumber = %v, want %v", unmarshaled.SerialNumber, tt.packet.SerialNumber)
			}

			// Compare TransferId (full UUID is stored)
			if unmarshaled.TransferId != tt.packet.TransferId {
				t.Errorf("TransferId = %v, want %v", unmarshaled.TransferId, tt.packet.TransferId)
			}

			// Compare timestamps (allowing for millisecond precision)
			expectedMillis := tt.packet.Timestamp.UnixMilli()
			actualMillis := unmarshaled.Timestamp.UnixMilli()
			if actualMillis != expectedMillis {
				t.Errorf("Timestamp = %v (millis: %d), want %v (millis: %d)",
					unmarshaled.Timestamp, actualMillis, tt.packet.Timestamp, expectedMillis)
			}

			if unmarshaled.DataSize != uint64(len(tt.packet.Data)) {
				t.Errorf("DataSize = %v, want %v", unmarshaled.DataSize, len(tt.packet.Data))
			}

			if !reflect.DeepEqual(unmarshaled.Data, tt.packet.Data) {
				t.Errorf("Data = %v, want %v", unmarshaled.Data, tt.packet.Data)
			}
		})
	}
}

func TestUdpPacket_Marshal_DataSize_Update(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)
	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 123,
		TransferId:   uuid.New(),
		Timestamp:    time.UnixMilli(1000001000),
		DataSize:     999, // Wrong value
		Data:         []byte("test"),
	}

	// Marshal should update DataSize
	_, err := packet.Marshal(transferStartTime)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if packet.DataSize != 4 {
		t.Errorf("DataSize should be updated to 4, got %d", packet.DataSize)
	}
}

func TestUdpPacket_Unmarshal_DataSize_Mismatch(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)
	// Create a packet with DataSize = 100 but only 10 bytes of actual data
	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 123,
		TransferId:   uuid.New(),
		Timestamp:    time.UnixMilli(1000001000),
		Data:         []byte("0123456789"), // 10 bytes
	}

	data, err := packet.Marshal(transferStartTime)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Manually modify DataSize in the marshaled data to be larger than available
	// DataSize is at offset: contentTypeSize(1) + serialNumberSize(4) + transferIdSize(16) + timestampSize(4) = 25
	dataSizeOffset := contentTypeSize + serialNumberSize + transferIdSize + timestampSize
	binary.BigEndian.PutUint16(data[dataSizeOffset:dataSizeOffset+dataSizeSize], 100) // Set to 100

	// Try to unmarshal - should fail
	unmarshaled := &UdpPacket{}
	err = unmarshaled.Unmarshal(data, transferStartTime)
	if err == nil {
		t.Error("Unmarshal() should fail when DataSize exceeds available data")
	}
	if err.Error() != "data too short: declared data size exceeds available data" {
		t.Errorf("Unmarshal() error = %v, want 'data too short: declared data size exceeds available data'", err)
	}
}

func TestUdpPacket_HeaderSize_Constant(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)
	// Verify header size is correct
	expectedHeaderSize := contentTypeSize + serialNumberSize + transferIdSize + timestampSize + dataSizeSize
	if headerSize != expectedHeaderSize {
		t.Errorf("headerSize = %d, want %d", headerSize, expectedHeaderSize)
	}

	// Verify header size matches actual marshaled header
	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 123,
		TransferId:   uuid.New(),
		Timestamp:    time.UnixMilli(1000001000),
		Data:         []byte{},
	}

	data, err := packet.Marshal(transferStartTime)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if len(data) != headerSize {
		t.Errorf("Marshaled data size = %d, want %d (header only)", len(data), headerSize)
	}
}

func TestUdpPacket_TransferId_Preservation(t *testing.T) {
	transferStartTime := time.UnixMilli(1000000000)
	// Test that TransferId is properly preserved (full UUID)
	testTransferId := uuid.New()

	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 123,
		TransferId:   testTransferId,
		Timestamp:    time.UnixMilli(1000001000),
		Data:         []byte("test data"),
	}

	data, err := packet.Marshal(transferStartTime)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	unmarshaled := &UdpPacket{}
	err = unmarshaled.Unmarshal(data, transferStartTime)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Compare full TransferId
	if unmarshaled.TransferId != testTransferId {
		t.Errorf("TransferId = %v, want %v", unmarshaled.TransferId, testTransferId)
	}
}

func BenchmarkUdpPacket_Marshal(b *testing.B) {
	transferStartTime := time.UnixMilli(1000000000)
	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 12345,
		TransferId:   uuid.New(),
		Timestamp:    time.UnixMilli(1000001000),
		Data:         make([]byte, 1024), // 1KB data
	}

	b.ResetTimer()
	for range b.N {
		_, _ = packet.Marshal(transferStartTime)
	}
}

func BenchmarkUdpPacket_Unmarshal(b *testing.B) {
	transferStartTime := time.UnixMilli(1000000000)
	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 12345,
		TransferId:   uuid.New(),
		Timestamp:    time.UnixMilli(1000001000),
		Data:         make([]byte, 1024), // 1KB data
	}

	data, err := packet.Marshal(transferStartTime)
	if err != nil {
		b.Fatalf("Marshal() error = %v", err)
	}

	b.ResetTimer()
	for range b.N {
		unmarshaled := &UdpPacket{}
		_ = unmarshaled.Unmarshal(data, transferStartTime)
	}
}

func BenchmarkUdpPacket_Marshal_Unmarshal_RoundTrip(b *testing.B) {
	transferStartTime := time.UnixMilli(1000000000)
	packet := &UdpPacket{
		ContentType:  0x01,
		SerialNumber: 12345,
		TransferId:   uuid.New(),
		Timestamp:    time.UnixMilli(1000001000),
		Data:         make([]byte, 1024), // 1KB data
	}

	b.ResetTimer()
	for range b.N {
		data, _ := packet.Marshal(transferStartTime)
		unmarshaled := &UdpPacket{}
		_ = unmarshaled.Unmarshal(data, transferStartTime)
	}
}
