package nfc

import (
	"fmt"
	
	"github.com/aarepuusepp/libopenpace-go/pace"
	"github.com/ebfe/scard"
)

// CardAdapter adapts scard.Card to libopenpace-go's CardInterface.
// This allows libopenpace-go to work with PC/SC smart card readers.
type CardAdapter struct {
	card *scard.Card
}

// NewCardAdapter creates a new CardAdapter wrapping an scard.Card.
func NewCardAdapter(card *scard.Card) pace.CardInterface {
	return &CardAdapter{card: card}
}

// Transmit sends an APDU command and returns the response.
// This implements the pace.CardInterface interface.
func (ca *CardAdapter) Transmit(command []byte) ([]byte, error) {
	if ca.card == nil {
		return nil, fmt.Errorf("card is nil")
	}
	
	response, err := ca.card.Transmit(command)
	if err != nil {
		return nil, fmt.Errorf("transmit failed: %w", err)
	}
	
	return response, nil
}

