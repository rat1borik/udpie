package utils

const (
	bitsPerByte  = 8
	maxBitInByte = 7
	byteMask     = 0xFF
)

type BitArray interface {
	Set(i uint64)
	Get(i uint64) bool
	And(offset uint64, other BitArray)
	Or(offset uint64, other BitArray)
	Neg()
	Clear(i uint64)
	Len() uint64
	Array() []byte
	Clone() BitArray
	SignificantPart() (uint64, BitArray)
}

type bitArray struct {
	array  []byte
	bitLen uint64 // actual bit count (from constructor argument)
}

// Создание битового массива на n бит
func NewBitArray(n uint64) BitArray {
	size := (n + maxBitInByte) / bitsPerByte // округляем вверх
	return &bitArray{
		array:  make([]byte, size),
		bitLen: n,
	}
}

func (arr *bitArray) Array() []byte {
	return arr.array
}

func (arr *bitArray) Len() uint64 {
	return arr.bitLen
}

func (arr *bitArray) And(offset uint64, other BitArray) {
	otherBitLen := other.Len()
	arrBitLen := arr.Len()

	if offset >= arrBitLen {
		return
	}

	bitOffset := offset % bitsPerByte
	byteOffset := offset / bitsPerByte

	otherArr := other.Array()
	otherByteLen := uint64(len(otherArr))
	arrByteLen := uint64(len(arr.array))

	// If byte-aligned, work with bytes directly for better performance
	if bitOffset == 0 {
		// Fast path: both are bitArray and byte-aligned - use direct byte operations
		bytesToProcess := (otherBitLen + maxBitInByte) / bitsPerByte
		if bytesToProcess > otherByteLen {
			bytesToProcess = otherByteLen
		}
		if byteOffset+bytesToProcess > arrByteLen {
			bytesToProcess = arrByteLen - byteOffset
		}

		// Direct byte-level AND operation
		//nolint:intrange // uint64 range requires explicit loop
		for i := uint64(0); i < bytesToProcess; i++ {
			arr.array[byteOffset+i] &= otherArr[i]
		}
	} else {
		// Non-byte-aligned: handle bit by bit but use direct array access
		// AND requires checking both bits, so bit-by-bit is necessary for correctness
		maxBits := otherBitLen
		if offset+maxBits > arrBitLen {
			maxBits = arrBitLen - offset
		}

		// Fast path: use direct array access for both
		otherBytes := otherArr
		otherByteCount := uint64(len(otherBytes))

		//nolint:intrange // uint64 range requires explicit loop
		for i := uint64(0); i < maxBits; i++ {
			arrBitIndex := offset + i
			otherByteIdx := i / bitsPerByte
			otherBitIdx := i % bitsPerByte

			if otherByteIdx >= otherByteCount {
				break
			}

			otherBit := (otherBytes[otherByteIdx] & (1 << otherBitIdx)) != 0
			byteIdx := arrBitIndex / bitsPerByte
			bitIdx := arrBitIndex % bitsPerByte
			mask := byte(1 << bitIdx)

			// Perform AND: if both are true, keep it set; otherwise clear it
			if (arr.array[byteIdx]&mask) != 0 && otherBit {
				arr.array[byteIdx] |= mask
			} else {
				arr.array[byteIdx] &^= mask
			}
		}
	}
}

//nolint:gocyclo // Complex bit manipulation logic requires multiple branches
func (arr *bitArray) Or(offset uint64, other BitArray) {
	otherBitLen := other.Len()
	arrBitLen := arr.Len()

	if offset >= arrBitLen {
		return
	}

	bitOffset := offset % bitsPerByte
	byteOffset := offset / bitsPerByte

	otherArr := other.Array()
	otherByteLen := uint64(len(otherArr))
	arrByteLen := uint64(len(arr.array))

	// If byte-aligned, work with bytes directly for better performance
	if bitOffset == 0 {
		// Fast path: byte-aligned - use direct byte operations
		bytesToProcess := (otherBitLen + maxBitInByte) / bitsPerByte
		if bytesToProcess > otherByteLen {
			bytesToProcess = otherByteLen
		}
		if byteOffset+bytesToProcess > arrByteLen {
			bytesToProcess = arrByteLen - byteOffset
		}

		// Direct byte-level OR operation
		//nolint:intrange // uint64 range requires explicit loop
		for i := uint64(0); i < bytesToProcess; i++ {
			arr.array[byteOffset+i] |= otherArr[i]
		}
	} else {
		// Non-byte-aligned: use bit shifting for better performance
		maxBits := otherBitLen
		if offset+maxBits > arrBitLen {
			maxBits = arrBitLen - offset
		}

		shift := bitOffset
		invShift := bitsPerByte - shift

		// Fast path: use direct array access with bit shifting
		otherBytes := otherArr
		otherByteCount := uint64(len(otherBytes))

		// Process whole bytes with bit shifting when possible
		bytesToProcess := otherByteCount
		if byteOffset+bytesToProcess >= arrByteLen {
			bytesToProcess = arrByteLen - byteOffset
		}

		//nolint:intrange // uint64 range requires explicit loop
		for i := uint64(0); i < bytesToProcess; i++ {
			otherByte := otherBytes[i]
			arrIdx1 := byteOffset + i

			if arrIdx1 < arrByteLen {
				// Lower bits: shift right into first byte
				masked := otherByte & ((1 << invShift) - 1)
				arr.array[arrIdx1] |= masked << shift
			}

			// Upper bits: shift left into next byte (if it exists)
			if shift > 0 && arrIdx1+1 < arrByteLen {
				masked := otherByte >> invShift
				arr.array[arrIdx1+1] |= masked
			}
		}

		// Handle remaining bits that don't fit in whole bytes
		processedBits := bytesToProcess * bitsPerByte
		if processedBits < maxBits {
			for i := processedBits; i < maxBits; i++ {
				arrBitIndex := offset + i
				if arrBitIndex >= arrBitLen {
					break
				}

				otherByteIdx := i / bitsPerByte
				otherBitIdx := i % bitsPerByte

				if otherByteIdx >= otherByteCount {
					break
				}

				if (otherBytes[otherByteIdx] & (1 << otherBitIdx)) != 0 {
					byteIdx := arrBitIndex / 8
					bitIdx := arrBitIndex % 8
					arr.array[byteIdx] |= 1 << bitIdx
				}
			}
		}
	}
}

func (arr *bitArray) Neg() {
	for i := range arr.array {
		arr.array[i] = ^arr.array[i]
	}
}

// Установить бит i
func (arr *bitArray) Set(i uint64) {
	if i >= arr.bitLen {
		return
	}
	arr.array[i/bitsPerByte] |= 1 << (i % bitsPerByte)
}

// Сбросить бит i
func (arr *bitArray) Clear(i uint64) {
	if i >= arr.bitLen {
		return
	}
	arr.array[i/8] &^= 1 << (i % 8)
}

// Проверить бит i
func (arr *bitArray) Get(i uint64) bool {
	if i >= arr.bitLen {
		return false
	}
	return (arr.array[i/bitsPerByte] & (1 << (i % bitsPerByte))) != 0
}

// Clone creates a deep copy of the bit array
func (arr *bitArray) Clone() BitArray {
	clonedArray := make([]byte, len(arr.array))
	copy(clonedArray, arr.array)
	return &bitArray{
		array:  clonedArray,
		bitLen: arr.bitLen,
	}
}

// SignificantPart returns the offset and a new BitArray containing only the significant part
// (from the first set bit to the last set bit, trimming leading and trailing zeros)
//
//nolint:gocyclo,funlen // Complex bit manipulation logic requires multiple branches and statements
func (arr *bitArray) SignificantPart() (uint64, BitArray) {
	arrByteLen := uint64(len(arr.array))
	arrBitLen := arr.Len()

	if arrByteLen == 0 || arrBitLen == 0 {
		return 0, NewBitArray(0)
	}

	// Find first non-zero byte and first set bit within it
	firstByteIdx := uint64(0)
	firstSetBit := uint64(0)
	foundFirst := false

	// Scan bytes to find first non-zero byte
	//nolint:intrange // uint64 range requires explicit loop
	for i := uint64(0); i < arrByteLen; i++ {
		if arr.array[i] != 0 {
			firstByteIdx = i
			// Find first set bit in this byte
			//nolint:intrange // uint64 range requires explicit loop
			for j := uint64(0); j < bitsPerByte; j++ {
				if (arr.array[i] & (1 << j)) != 0 {
					firstSetBit = i*bitsPerByte + j
					foundFirst = true
					break
				}
			}
			break
		}
	}

	// If no bits are set, return offset 0 and empty array
	if !foundFirst {
		return 0, NewBitArray(0)
	}

	// Find last non-zero byte and last set bit within it
	lastSetBit := firstSetBit

	// Scan bytes from the end to find last non-zero byte
	// But don't go beyond the actual bit length
	maxByteToCheck := (arrBitLen - 1) / bitsPerByte
	if maxByteToCheck >= arrByteLen {
		maxByteToCheck = arrByteLen - 1
	}

	for i := maxByteToCheck; i >= firstByteIdx; i-- {
		if arr.array[i] != 0 {
			// Find last set bit in this byte, but don't go beyond bitLen
			maxBitInByteVal := uint64(maxBitInByte)
			if i == maxByteToCheck {
				maxBitInByteVal = (arrBitLen - 1) % bitsPerByte
			}
			//nolint:gosec // maxBitInByteVal is bounded by bitsPerByte-1, safe to convert
			for j := int(maxBitInByteVal); j >= 0; j-- {
				bitPos := uint64(j)
				if (arr.array[i] & (1 << bitPos)) != 0 {
					lastSetBit = i*bitsPerByte + bitPos
					break
				}
			}
			break
		}
	}

	// Calculate the size needed for the significant part
	significantBitLen := lastSetBit - firstSetBit + 1
	significantArray := NewBitArray(significantBitLen)
	sigArr := significantArray.(*bitArray)

	// Calculate byte boundaries
	firstByteOffset := firstSetBit / bitsPerByte
	firstBitOffset := firstSetBit % bitsPerByte
	lastByteOffset := lastSetBit / bitsPerByte

	// Copy bytes efficiently
	if firstBitOffset == 0 {
		// Byte-aligned: can copy whole bytes
		bytesToCopy := lastByteOffset - firstByteOffset + 1
		sigBytesNeeded := (significantBitLen + maxBitInByte) / bitsPerByte

		if bytesToCopy > sigBytesNeeded {
			bytesToCopy = sigBytesNeeded
		}

		//nolint:intrange // uint64 range requires explicit loop
		for i := uint64(0); i < bytesToCopy; i++ {
			if firstByteOffset+i < arrByteLen {
				sigArr.array[i] = arr.array[firstByteOffset+i]
			}
		}

		// Clear trailing bits if needed
		if significantBitLen%8 != 0 {
			clearMask := byte(byteMask) << (significantBitLen % bitsPerByte)
			sigArr.array[sigBytesNeeded-1] &^= clearMask
		}
	} else {
		// Non-byte-aligned: need to shift bits
		shift := firstBitOffset
		invShift := 8 - shift

		sigBytesNeeded := (significantBitLen + maxBitInByte) / bitsPerByte

		//nolint:intrange // uint64 range requires explicit loop
		for i := uint64(0); i < sigBytesNeeded; i++ {
			srcByteIdx := firstByteOffset + i

			// Lower bits from current source byte (shift right to align)
			if srcByteIdx < arrByteLen {
				// Extract bits starting from firstBitOffset
				masked := arr.array[srcByteIdx] >> shift
				sigArr.array[i] = masked
			}

			// Upper bits from next source byte (if exists and needed)
			if shift > 0 && (srcByteIdx+1) < arrByteLen {
				// Extract upper bits from next byte
				masked := arr.array[srcByteIdx+1] & ((1 << shift) - 1)
				sigArr.array[i] |= masked << invShift
			}
		}

		// Clear trailing bits if needed
		if significantBitLen%bitsPerByte != 0 {
			clearMask := byte(byteMask) << (significantBitLen % bitsPerByte)
			sigArr.array[sigBytesNeeded-1] &^= clearMask
		}
	}

	return firstSetBit, significantArray
}
