package esteid

import (
	"bytes"
	"strings"
)

// APDU helper functions for reuse in both contact and NFC modes

// APDUSelectDF creates a SELECT DF command.
func APDUSelectDF(fid [2]byte) []byte {
	return []byte{0x00, 0xA4, 0x01, 0x0C, 0x02, fid[0], fid[1]}
}

// APDUSelectEF creates a SELECT EF command.
func APDUSelectEF(fid [2]byte) []byte {
	return []byte{0x00, 0xA4, 0x02, 0x0C, 0x02, fid[0], fid[1]}
}

// APDUSelectByFID creates a SELECT by FID command.
func APDUSelectByFID(fid [2]byte) []byte {
	return []byte{0x00, 0xA4, 0x00, 0x0C, 0x02, fid[0], fid[1]}
}

// APDUSelectAID creates a SELECT by AID command.
func APDUSelectAID(aid []byte) []byte {
	cmd := make([]byte, 5+len(aid))
	cmd[0] = 0x00 // CLA
	cmd[1] = 0xA4 // INS
	cmd[2] = 0x04 // P1 (select by AID)
	cmd[3] = 0x00 // P2
	cmd[4] = byte(len(aid))
	copy(cmd[5:], aid)
	return cmd
}

// APDUReadBinary creates a READ BINARY command.
// If le <= 0, uses Le=0 (256 bytes or until 6Cxx).
func APDUReadBinary(offset int, le int) []byte {
	p1 := byte((offset >> 8) & 0xFF)
	p2 := byte(offset & 0xFF)
	if le <= 0 {
		return []byte{0x00, 0xB0, p1, p2, 0x00}
	}
	if le > 255 {
		le = 255
	}
	return []byte{0x00, 0xB0, p1, p2, byte(le)}
}

// BytesToCleanText converts raw bytes to clean UTF-8 text.
// Strips trailing nulls, spaces, and control characters.
func BytesToCleanText(b []byte) string {
	// Trim trailing 0x00, 0x20, \n, \r
	trim := b
	for len(trim) > 0 {
		last := trim[len(trim)-1]
		if last == 0x00 || last == 0x20 || last == '\n' || last == '\r' {
			trim = trim[:len(trim)-1]
		} else {
			break
		}
	}

	// Remove internal NULs
	out := make([]rune, 0, len(trim))
	for _, c := range string(trim) {
		if c != '\x00' {
			out = append(out, c)
		}
	}

	return strings.TrimSpace(string(out))
}

// SanitizeTextBytes is an alias for BytesToCleanText for compatibility.
func SanitizeTextBytes(b []byte) string {
	return BytesToCleanText(b)
}

// ParseSW extracts status word from response.
func ParseSW(resp []byte) (sw1, sw2 byte, data []byte) {
	if len(resp) < 2 {
		return 0, 0, resp
	}
	return resp[len(resp)-2], resp[len(resp)-1], resp[:len(resp)-2]
}

// CombineSW combines SW1 and SW2 into a single uint16.
func CombineSW(sw1, sw2 byte) uint16 {
	return uint16(sw1)<<8 | uint16(sw2)
}

// IsSuccess checks if SW indicates success (9000).
func IsSuccess(sw uint16) bool {
	return sw == 0x9000
}

// IsMoreData checks if SW indicates more data available (61xx).
func IsMoreData(sw uint16) bool {
	return (sw >> 8) == 0x61
}

// IsWrongLength checks if SW indicates wrong length (6Cxx).
func IsWrongLength(sw uint16) bool {
	return (sw >> 8) == 0x6C
}

// GetCorrectLength extracts the correct length from 6Cxx response.
func GetCorrectLength(sw uint16) int {
	if IsWrongLength(sw) {
		return int(sw & 0x00FF)
	}
	return 0
}

// TrimBinaryPadding removes trailing padding from binary data.
func TrimBinaryPadding(data []byte) []byte {
	return bytes.TrimRight(data, "\x00\x20\n\r")
}

