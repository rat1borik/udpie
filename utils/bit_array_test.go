package utils

import (
	"fmt"
	"testing"
)

func TestBitArray_And(t *testing.T) {
	tests := []struct {
		name            string
		arrSize         uint64
		otherSize       uint64
		arrBits         []uint64 // bits to set in arr
		otherBits       []uint64 // bits to set in other
		offset          uint64
		expectedBits    []uint64 // bits that should be set after AND
		expectedCleared []uint64 // bits that should be cleared after AND
	}{
		{
			name:            "simple AND at offset 0",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{0, 1, 2, 3},
			otherBits:       []uint64{1, 2, 3},
			offset:          0,
			expectedBits:    []uint64{1, 2, 3},
			expectedCleared: []uint64{0},
		},
		{
			name:            "AND with offset",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{5, 6, 7, 8},
			otherBits:       []uint64{0, 1, 2},
			offset:          5,
			expectedBits:    []uint64{5, 6, 7},
			expectedCleared: []uint64{8},
		},
		{
			name:            "AND with no overlap",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{0, 1},
			otherBits:       []uint64{2, 3},
			offset:          0,
			expectedBits:    []uint64{},
			expectedCleared: []uint64{0, 1},
		},
		{
			name:            "AND with partial overlap",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{0, 1, 2, 3, 4},
			otherBits:       []uint64{1, 3, 5},
			offset:          0,
			expectedBits:    []uint64{1, 3},
			expectedCleared: []uint64{0, 2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := NewBitArray(tt.arrSize)
			other := NewBitArray(tt.otherSize)

			// Set bits in arr
			for _, bit := range tt.arrBits {
				arr.Set(bit)
			}

			// Set bits in other
			for _, bit := range tt.otherBits {
				other.Set(bit)
			}

			// Perform AND operation
			arr.And(tt.offset, other)

			// Check expected bits are set
			for _, bit := range tt.expectedBits {
				if !arr.Get(bit) {
					t.Errorf("Expected bit %d to be set after AND, but it was not", bit)
				}
			}

			// Check expected bits are cleared
			for _, bit := range tt.expectedCleared {
				if arr.Get(bit) {
					t.Errorf("Expected bit %d to be cleared after AND, but it was set", bit)
				}
			}
		})
	}
}

func TestBitArray_Or(t *testing.T) {
	tests := []struct {
		name            string
		arrSize         uint64
		otherSize       uint64
		arrBits         []uint64 // bits to set in arr
		otherBits       []uint64 // bits to set in other
		offset          uint64
		expectedBits    []uint64 // bits that should be set after OR
		expectedCleared []uint64 // bits that should remain cleared after OR
	}{
		{
			name:            "simple OR at offset 0",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{0, 1},
			otherBits:       []uint64{2, 3},
			offset:          0,
			expectedBits:    []uint64{0, 1, 2, 3},
			expectedCleared: []uint64{4, 5, 6, 7},
		},
		{
			name:            "OR with offset",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{0, 1},
			otherBits:       []uint64{0, 1, 2},
			offset:          5,
			expectedBits:    []uint64{0, 1, 5, 6, 7},
			expectedCleared: []uint64{2, 3, 4, 8, 9},
		},
		{
			name:            "OR with no initial bits",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{},
			otherBits:       []uint64{0, 1, 2},
			offset:          0,
			expectedBits:    []uint64{0, 1, 2},
			expectedCleared: []uint64{3, 4, 5, 6, 7},
		},
		{
			name:            "OR with overlapping bits",
			arrSize:         16,
			otherSize:       8,
			arrBits:         []uint64{0, 2, 4},
			otherBits:       []uint64{1, 2, 3},
			offset:          0,
			expectedBits:    []uint64{0, 1, 2, 3, 4},
			expectedCleared: []uint64{5, 6, 7},
		},
		{
			name:            "OR with large offset",
			arrSize:         32,
			otherSize:       8,
			arrBits:         []uint64{0},
			otherBits:       []uint64{0, 1, 2},
			offset:          10,
			expectedBits:    []uint64{0, 10, 11, 12},
			expectedCleared: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14, 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := NewBitArray(tt.arrSize)
			other := NewBitArray(tt.otherSize)

			// Set bits in arr
			for _, bit := range tt.arrBits {
				arr.Set(bit)
			}

			// Set bits in other
			for _, bit := range tt.otherBits {
				other.Set(bit)
			}

			// Perform OR operation
			arr.Or(tt.offset, other)

			// Check expected bits are set
			for _, bit := range tt.expectedBits {
				if !arr.Get(bit) {
					t.Errorf("Expected bit %d to be set after OR, but it was not", bit)
				}
			}

			// Check expected bits remain cleared
			for _, bit := range tt.expectedCleared {
				if arr.Get(bit) {
					t.Errorf("Expected bit %d to remain cleared after OR, but it was set", bit)
				}
			}
		})
	}
}

func TestBitArray_And_Or_Combined(t *testing.T) {
	// Test that AND and OR work together correctly
	arr := NewBitArray(16)
	other1 := NewBitArray(8)
	other2 := NewBitArray(8)

	// Set some bits in arr
	arr.Set(0)
	arr.Set(1)
	arr.Set(2)

	// Set bits in other1
	other1.Set(1)
	other1.Set(2)
	other1.Set(3)

	// Set bits in other2
	other2.Set(0)
	other2.Set(2)

	// First AND: arr AND other1 at offset 0
	// Result should be: bits 1, 2 set (common bits)
	arr.And(0, other1)
	if !arr.Get(1) || !arr.Get(2) {
		t.Error("After AND with other1, bits 1 and 2 should be set")
	}
	if arr.Get(0) || arr.Get(3) {
		t.Error("After AND with other1, bits 0 and 3 should be cleared")
	}

	// Then OR: arr OR other2 at offset 0
	// Result should be: bits 0, 1, 2 set
	arr.Or(0, other2)
	if !arr.Get(0) || !arr.Get(1) || !arr.Get(2) {
		t.Error("After OR with other2, bits 0, 1, and 2 should be set")
	}
}

func TestBitArray_And_Or_Boundary(t *testing.T) {
	// Test boundary conditions
	arr := NewBitArray(16)
	other := NewBitArray(8)

	// Set all bits in other
	for i := range uint64(8) {
		other.Set(i)
	}

	// OR at offset that would go beyond arr bounds
	arr.Or(10, other)

	// Check that bits 10-15 are set (within bounds)
	for i := uint64(10); i < 16; i++ {
		if !arr.Get(i) {
			t.Errorf("Expected bit %d to be set after OR, but it was not", i)
		}
	}

	// Check that bits 0-9 remain cleared
	for i := range uint64(10) {
		if arr.Get(i) {
			t.Errorf("Expected bit %d to remain cleared, but it was set", i)
		}
	}
}

func TestBitArray_Clone(t *testing.T) {
	tests := []struct {
		name    string
		size    uint64
		setBits []uint64
	}{
		{
			name:    "clone empty array",
			size:    16,
			setBits: []uint64{},
		},
		{
			name:    "clone with some bits set",
			size:    16,
			setBits: []uint64{0, 1, 5, 10, 15},
		},
		{
			name:    "clone with all bits set",
			size:    8,
			setBits: []uint64{0, 1, 2, 3, 4, 5, 6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewBitArray(tt.size)

			// Set bits in original
			for _, bit := range tt.setBits {
				original.Set(bit)
			}

			// Clone the array
			cloned := original.Clone()

			// Verify all bits match
			//nolint:intrange // uint64 range requires explicit loop
			for i := uint64(0); i < tt.size; i++ {
				originalBit := original.Get(i)
				clonedBit := cloned.Get(i)
				if originalBit != clonedBit {
					t.Errorf("Bit %d: original=%v, cloned=%v (mismatch)", i, originalBit, clonedBit)
				}
			}

			// Verify they are independent - modify original and check cloned is unchanged
			if len(tt.setBits) > 0 {
				// Clear a bit in original
				original.Clear(tt.setBits[0])
				if cloned.Get(tt.setBits[0]) != original.Get(tt.setBits[0]) {
					// This is expected - they should be different now
					if !cloned.Get(tt.setBits[0]) {
						t.Errorf("Cloned array was modified when original was changed")
					}
				}
			}
		})
	}
}

//nolint:gocyclo // Test function with many test cases
func TestBitArray_SignificantPart(t *testing.T) {
	tests := []struct {
		name           string
		size           uint64
		setBits        []uint64
		expectedOffset uint64
		expectedBits   []uint64 // bits that should be set in the significant part (relative to offset)
	}{
		{
			name:           "empty array",
			size:           16,
			setBits:        []uint64{},
			expectedOffset: 0,
			expectedBits:   []uint64{},
		},
		{
			name:           "single bit at start",
			size:           16,
			setBits:        []uint64{0},
			expectedOffset: 0,
			expectedBits:   []uint64{0},
		},
		{
			name:           "bits at start",
			size:           16,
			setBits:        []uint64{0, 1, 2},
			expectedOffset: 0,
			expectedBits:   []uint64{0, 1, 2},
		},
		{
			name:           "bits with leading zeros",
			size:           16,
			setBits:        []uint64{5, 6, 7},
			expectedOffset: 5,
			expectedBits:   []uint64{0, 1, 2},
		},
		{
			name:           "bits with trailing zeros",
			size:           16,
			setBits:        []uint64{0, 1, 2, 3},
			expectedOffset: 0,
			expectedBits:   []uint64{0, 1, 2, 3},
		},
		{
			name:           "bits with both leading and trailing zeros",
			size:           32,
			setBits:        []uint64{5, 6, 7, 8, 9},
			expectedOffset: 5,
			expectedBits:   []uint64{0, 1, 2, 3, 4},
		},
		{
			name:           "sparse bits",
			size:           32,
			setBits:        []uint64{3, 10, 20},
			expectedOffset: 3,
			expectedBits:   []uint64{0, 7, 17},
		},
		{
			name:           "single bit in middle",
			size:           16,
			setBits:        []uint64{10},
			expectedOffset: 10,
			expectedBits:   []uint64{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := NewBitArray(tt.size)

			// Set bits
			for _, bit := range tt.setBits {
				arr.Set(bit)
			}

			// Get significant part
			offset, significant := arr.SignificantPart()

			// Check offset
			if offset != tt.expectedOffset {
				t.Errorf("Expected offset %d, got %d", tt.expectedOffset, offset)
			}

			// Check significant part size
			expectedSize := uint64(0)
			if len(tt.expectedBits) > 0 {
				maxBit := uint64(0)
				for _, bit := range tt.expectedBits {
					if bit > maxBit {
						maxBit = bit
					}
				}
				expectedSize = maxBit + 1
			}

			if significant.Len() < expectedSize {
				t.Errorf("Significant part too small: expected at least %d bits, got %d", expectedSize, significant.Len())
			}

			// Check that all expected bits are set
			for _, relBit := range tt.expectedBits {
				if !significant.Get(relBit) {
					t.Errorf("Expected bit %d (relative) to be set in significant part", relBit)
				}
			}

			// Verify the significant part matches the original when offset is applied
			// Only check up to the actual significant bit length
			significantBitLen := uint64(0)
			if len(tt.expectedBits) > 0 {
				maxBit := uint64(0)
				for _, bit := range tt.expectedBits {
					if bit > maxBit {
						maxBit = bit
					}
				}
				significantBitLen = maxBit + 1
			}

			for i := uint64(0); i < significantBitLen && i < significant.Len(); i++ {
				sigBit := significant.Get(i)
				origBit := arr.Get(offset + i)
				if sigBit != origBit {
					t.Errorf("Mismatch at relative bit %d (absolute %d): significant=%v, original=%v",
						i, offset+i, sigBit, origBit)
				}
			}
		})
	}
}

func TestBitArray_Clone_Independence(t *testing.T) {
	// Test that cloned arrays are truly independent
	original := NewBitArray(16)
	original.Set(0)
	original.Set(5)
	original.Set(10)

	cloned := original.Clone()

	// Modify original
	original.Clear(5)
	original.Set(7)

	// Verify cloned is unchanged
	if !cloned.Get(0) {
		t.Error("Cloned bit 0 should still be set")
	}
	if !cloned.Get(5) {
		t.Error("Cloned bit 5 should still be set (original was cleared)")
	}
	if !cloned.Get(10) {
		t.Error("Cloned bit 10 should still be set")
	}
	if cloned.Get(7) {
		t.Error("Cloned bit 7 should not be set (only original was set)")
	}

	// Verify original is changed
	if original.Get(5) {
		t.Error("Original bit 5 should be cleared")
	}
	if !original.Get(7) {
		t.Error("Original bit 7 should be set")
	}
}

//nolint:gocyclo // Stress test function with many test cases
func TestBitArray_SignificantPart_Stress(t *testing.T) {
	// Stress test with large arrays
	t.Run("large_array_sparse", func(t *testing.T) {
		size := uint64(10000)
		arr := NewBitArray(size)

		// Set bits at the beginning, middle, and end
		arr.Set(0)
		arr.Set(5000)
		arr.Set(9999)

		offset, significant := arr.SignificantPart()

		if offset != 0 {
			t.Errorf("Expected offset 0, got %d", offset)
		}

		// Should contain bits 0, 5000, and 9999
		if !significant.Get(0) {
			t.Error("Bit 0 should be set")
		}
		if !significant.Get(5000) {
			t.Error("Bit 5000 should be set")
		}
		if !significant.Get(9999) {
			t.Error("Bit 9999 should be set")
		}

		// Verify size is correct (should be 10000 bits)
		if significant.Len() < 10000 {
			t.Errorf("Significant part too small: expected at least 10000 bits, got %d", significant.Len())
		}
	})

	t.Run("large_array_dense", func(t *testing.T) {
		size := uint64(10000)
		arr := NewBitArray(size)

		// Set every 10th bit
		for i := uint64(0); i < size; i += 10 {
			arr.Set(i)
		}

		offset, significant := arr.SignificantPart()

		if offset != 0 {
			t.Errorf("Expected offset 0, got %d", offset)
		}

		// Verify all set bits are present
		for i := uint64(0); i < size; i += 10 {
			if !significant.Get(i) {
				t.Errorf("Bit %d should be set in significant part", i)
			}
		}
	})

	t.Run("large_array_with_leading_zeros", func(t *testing.T) {
		size := uint64(10000)
		arr := NewBitArray(size)

		// Set bits starting from 5000
		startBit := uint64(5000)
		for i := startBit; i < startBit+100; i++ {
			arr.Set(i)
		}

		offset, significant := arr.SignificantPart()

		if offset != startBit {
			t.Errorf("Expected offset %d, got %d", startBit, offset)
		}

		// Verify all bits are correctly copied
		for i := range uint64(100) {
			if !significant.Get(i) {
				t.Errorf("Relative bit %d (absolute %d) should be set", i, startBit+i)
			}
		}
	})

	t.Run("very_large_array", func(t *testing.T) {
		size := uint64(100000)
		arr := NewBitArray(size)

		// Set bits at strategic positions
		arr.Set(0)
		arr.Set(1000)
		arr.Set(50000)
		arr.Set(99999)

		offset, significant := arr.SignificantPart()

		if offset != 0 {
			t.Errorf("Expected offset 0, got %d", offset)
		}

		// Verify all set bits are present
		if !significant.Get(0) {
			t.Error("Bit 0 should be set")
		}
		if !significant.Get(1000) {
			t.Error("Bit 1000 should be set")
		}
		if !significant.Get(50000) {
			t.Error("Bit 50000 should be set")
		}
		if !significant.Get(99999) {
			t.Error("Bit 99999 should be set")
		}
	})

	t.Run("non_byte_aligned_stress", func(t *testing.T) {
		// Test various non-byte-aligned offsets
		for _, startBit := range []uint64{1, 3, 5, 7, 9, 15, 17} {
			t.Run(fmt.Sprintf("offset_%d", startBit), func(t *testing.T) {
				size := uint64(1000)
				arr := NewBitArray(size)

				// Set a range of bits starting at non-byte-aligned position
				for i := startBit; i < startBit+50; i++ {
					arr.Set(i)
				}

				offset, significant := arr.SignificantPart()

				if offset != startBit {
					t.Errorf("Expected offset %d, got %d", startBit, offset)
				}

				// Verify all bits are correctly copied
				for i := range uint64(50) {
					expected := arr.Get(startBit + i)
					actual := significant.Get(i)
					if expected != actual {
						t.Errorf("Mismatch at relative bit %d (absolute %d): expected=%v, actual=%v",
							i, startBit+i, expected, actual)
					}
				}
			})
		}
	})

	t.Run("single_byte_stress", func(t *testing.T) {
		// Test with arrays that fit in a single byte
		for size := uint64(1); size <= 8; size++ {
			t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
				arr := NewBitArray(size)

				// Set all bits
				//nolint:intrange // uint64 range requires explicit loop
				//nolint:intrange // uint64 range requires explicit loop
				for i := uint64(0); i < size; i++ {
					arr.Set(i)
				}

				offset, significant := arr.SignificantPart()

				if offset != 0 {
					t.Errorf("Expected offset 0, got %d", offset)
				}

				// Verify all bits are set
				//nolint:intrange // uint64 range requires explicit loop
				//nolint:intrange // uint64 range requires explicit loop
				for i := uint64(0); i < size; i++ {
					if !significant.Get(i) {
						t.Errorf("Bit %d should be set", i)
					}
				}
			})
		}
	})

	t.Run("edge_cases", func(t *testing.T) {
		// Test edge cases
		testCases := []struct {
			name    string
			size    uint64
			setBits []uint64
		}{
			{"last_bit_only", 100, []uint64{99}},
			{"first_and_last", 100, []uint64{0, 99}},
			{"middle_range", 100, []uint64{45, 46, 47, 48, 49, 50}},
			{"alternating", 100, []uint64{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				arr := NewBitArray(tc.size)
				for _, bit := range tc.setBits {
					arr.Set(bit)
				}

				offset, significant := arr.SignificantPart()

				// Verify all set bits are in the significant part
				for _, bit := range tc.setBits {
					relBit := bit - offset
					if relBit < significant.Len() {
						if !significant.Get(relBit) {
							t.Errorf("Bit %d (relative %d) should be set in significant part", bit, relBit)
						}
					}
				}

				// Verify significant part matches original when offset is applied
				for i := uint64(0); i < significant.Len() && (offset+i) < tc.size; i++ {
					expected := arr.Get(offset + i)
					actual := significant.Get(i)
					if expected != actual {
						t.Errorf("Mismatch at relative bit %d (absolute %d): expected=%v, actual=%v",
							i, offset+i, expected, actual)
					}
				}
			})
		}
	})
}

func BenchmarkBitArray_SignificantPart(b *testing.B) {
	sizes := []uint64{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d_sparse", size), func(b *testing.B) {
			arr := NewBitArray(size)
			// Set bits at beginning, middle, and end
			arr.Set(0)
			arr.Set(size / 2)
			arr.Set(size - 1)

			b.ResetTimer()
			for range b.N {
				_, _ = arr.SignificantPart()
			}
		})

		b.Run(fmt.Sprintf("size_%d_dense", size), func(b *testing.B) {
			arr := NewBitArray(size)
			// Set every 10th bit
			for i := uint64(0); i < size; i += 10 {
				arr.Set(i)
			}

			b.ResetTimer()
			for range b.N {
				_, _ = arr.SignificantPart()
			}
		})

		b.Run(fmt.Sprintf("size_%d_range", size), func(b *testing.B) {
			arr := NewBitArray(size)
			// Set a range of bits
			start := size / 4
			for i := start; i < start+size/2; i++ {
				arr.Set(i)
			}

			b.ResetTimer()
			for range b.N {
				_, _ = arr.SignificantPart()
			}
		})
	}
}
