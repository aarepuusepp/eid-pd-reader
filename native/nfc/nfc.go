package nfc

import (
	"fmt"
	"strings"

	"github.com/ebfe/scard"
)

// ConnectAny connects to the first available NFC/contactless reader.
// Returns PC/SC context, card connection, and reader name.
func ConnectAny() (*scard.Context, *scard.Card, string, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, nil, "", fmt.Errorf("pcsc: establish context: %w", err)
	}

	rdrs, err := ctx.ListReaders()
	if err != nil {
		ctx.Release()
		return nil, nil, "", fmt.Errorf("pcsc: list readers: %w", err)
	}

	if len(rdrs) == 0 {
		ctx.Release()
		return nil, nil, "", fmt.Errorf("no readers found")
	}

	// Try to connect to NFC/contactless readers first
	for _, r := range rdrs {
		if IsContactlessReader(r) {
			card, err := ctx.Connect(r, scard.ShareShared, scard.ProtocolAny)
			if err == nil {
				return ctx, card, r, nil
			}
		}
	}

	// If no contactless reader found, try any reader
	for _, r := range rdrs {
		card, err := ctx.Connect(r, scard.ShareShared, scard.ProtocolAny)
		if err == nil {
			return ctx, card, r, nil
		}
	}

	ctx.Release()
	return nil, nil, "", fmt.Errorf("no NFC reader available or no card present")
}

// IsContactlessReader checks if a reader name suggests it's a contactless/NFC reader.
func IsContactlessReader(readerName string) bool {
	nameLower := strings.ToLower(readerName)
	return strings.Contains(nameLower, "cl reader") ||
		strings.Contains(nameLower, "picc") ||
		strings.Contains(nameLower, "nfc") ||
		strings.Contains(nameLower, "contactless")
}

// IsContactless checks if the card/reader combination is contactless.
func IsContactless(card *scard.Card, readerName string) bool {
	// Primary check: reader name
	if IsContactlessReader(readerName) {
		return true
	}

	// Secondary check: ATR/ATS analysis (if needed in future)
	// For now, rely on reader name
	return false
}

