package nfc

import (
	"fmt"
)

const (
	// MaxSMLe is the maximum Le (expected length) for SM-wrapped APDUs.
	// 223 plaintext bytes pad to 224; DO87 + DO99 + DO8E then fit in 255.
	MaxSMLe = 223
)

// SMReadBinary reads binary data from the currently selected EF via Secure Messaging.
// Handles 6Cxx (wrong length) responses and chunking for large files.
func SMReadBinary(sm SM, offset int, maxLength int) ([]byte, error) {
	if maxLength <= 0 {
		maxLength = 65536 // reasonable max
	}

	var result []byte
	currentOffset := offset

	for len(result) < maxLength {
		// Determine how much to read in this chunk
		remaining := maxLength - len(result)
		chunkSize := MaxSMLe
		if remaining < chunkSize {
			chunkSize = remaining
		}

		apdu := APDUReadBinary(currentOffset, chunkSize)
		data, sw, err := sm.Transmit(apdu)
		if err != nil {
			return nil, fmt.Errorf("SM transmit failed at offset %d: %w", currentOffset, err)
		}

		// Handle 6Cxx - wrong length, retry with correct Le
		if IsWrongLength(sw) {
			correctLen := GetCorrectLength(sw)
			if correctLen > 0 && correctLen <= chunkSize {
				apdu = APDUReadBinary(currentOffset, correctLen)
				data, sw, err = sm.Transmit(apdu)
				if err != nil {
					return nil, fmt.Errorf("SM transmit retry failed at offset %d: %w", currentOffset, err)
				}
			}
		}
		if len(data) > chunkSize {
			return nil, fmt.Errorf("READ BINARY exceeded requested length")
		}

		// Check for success or EOF warning
		if sw == 0x6282 {
			// "End of file reached before reading Le bytes" — append any data and stop
			if len(data) > 0 {
				result = append(result, data...)
			}
			break
		}
		if !IsSuccess(sw) {
			if len(result) > 0 && (sw == 0x6B00 || sw == 0x6A82) {
				break
			}
			return nil, fmt.Errorf("READ BINARY failed at offset %d: SW=%04X", currentOffset, sw)
		}

		// Append data
		if len(data) == 0 {
			break // No more data
		}

		result = append(result, data...)
		currentOffset += len(data)

		// If we got less than requested, we're done
		if len(data) < chunkSize {
			break
		}
	}

	return result, nil
}

// SMReadEF reads an entire EF via Secure Messaging.
// Uses bounded short APDUs that fit inside AES secure messaging.
func SMReadEF(sm SM, maxSize int) ([]byte, error) {
	if maxSize <= 0 {
		maxSize = 512 // default for personal data fields
	}

	return SMReadBinary(sm, 0, maxSize)
}

// SMSelectAndRead selects an EF and reads its content via Secure Messaging.
func SMSelectAndRead(sm SM, fid [2]byte, maxSize int) ([]byte, error) {
	// Select EF
	selectAPDU := APDUSelectEF(fid)
	_, sw, err := sm.Transmit(selectAPDU)
	if err != nil {
		return nil, fmt.Errorf("SELECT EF %02X%02X failed: %w", fid[0], fid[1], err)
	}

	if !IsSuccess(sw) {
		return nil, fmt.Errorf("SELECT EF %02X%02X returned SW=%04X", fid[0], fid[1], sw)
	}

	// Read EF content
	data, err := SMReadEF(sm, maxSize)
	if err != nil {
		return nil, fmt.Errorf("READ EF %02X%02X failed: %w", fid[0], fid[1], err)
	}

	return data, nil
}
