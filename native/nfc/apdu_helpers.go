package nfc

// APDU helper functions for NFC communication.
// These are duplicated from esteid package to avoid import cycles.

// APDUReadBinary creates a READ BINARY APDU command.
func APDUReadBinary(offset int, le int) []byte {
	p1 := byte((offset >> 8) & 0xFF)
	p2 := byte(offset & 0xFF)
	
	if le == 0 {
		return []byte{0x00, 0xB0, p1, p2, 0x00}
	}
	return []byte{0x00, 0xB0, p1, p2, byte(le)}
}

// APDUSelectEF creates a SELECT FILE (EF) APDU command.
func APDUSelectEF(fid [2]byte) []byte {
	return []byte{0x00, 0xA4, 0x02, 0x0C, 0x02, fid[0], fid[1]}
}

// IsSuccess checks if a status word indicates success (0x9000).
func IsSuccess(sw uint16) bool {
	return sw == 0x9000
}

// IsWrongLength checks if a status word indicates wrong length (0x6Cxx).
func IsWrongLength(sw uint16) bool {
	return (sw & 0xFF00) == 0x6C00
}

// GetCorrectLength extracts the correct length from a 0x6Cxx status word.
func GetCorrectLength(sw uint16) int {
	return int(sw & 0x00FF)
}

