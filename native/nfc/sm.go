package nfc

// SM (Secure Messaging) interface abstracts PACE-protected communication with the card.
// After PACE is established, all APDUs must be wrapped with encryption and MAC.
type SM interface {
	// Transmit sends a plain APDU, wraps it with SM, sends to card, unwraps response.
	// Returns the unwrapped response data and status word (SW).
	Transmit(apdu []byte) (data []byte, sw uint16, err error)
}

