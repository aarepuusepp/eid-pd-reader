package esteid

import (
	"fmt"

	"github.com/ebfe/scard"
)

// CardMonitor wraps PC/SC reader presence checks for EstEID cards.
type CardMonitor struct{}

func NewCardMonitor() *CardMonitor {
	return &CardMonitor{}
}

// RequireCardPresent returns an error when no smart card is present in any reader.
func (m *CardMonitor) RequireCardPresent() error {
	if m == nil {
		return fmt.Errorf("CARD_NOT_FOUND: card monitor unavailable")
	}
	ctx, err := scard.EstablishContext()
	if err != nil {
		return fmt.Errorf("CARD_NOT_FOUND: failed to establish PC/SC context: %w", err)
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		return fmt.Errorf("CARD_NOT_FOUND: failed to list readers: %w", err)
	}
	if len(readers) == 0 {
		return fmt.Errorf("CARD_NOT_FOUND: no card readers detected")
	}

	for _, reader := range readers {
		card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			continue
		}
		_ = card.Disconnect(scard.LeaveCard)
		return nil
	}
	return fmt.Errorf("CARD_NOT_FOUND: insert your ID card into the reader")
}
