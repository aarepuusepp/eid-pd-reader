package nfc

import (
	"fmt"

	"github.com/aarepuusepp/libopenpace-go/sm"
	"github.com/ebfe/scard"
)

// smPureGo implements the SM interface using libopenpace-go.
type smPureGo struct {
	card  *scard.Card
	smCtx *sm.Context
}

// NewSMPureGo creates a new Secure Messaging wrapper using libopenpace-go.
func NewSMPureGo(card *scard.Card, smCtx *sm.Context) SM {
	return &smPureGo{
		card:  card,
		smCtx: smCtx,
	}
}

// Transmit sends a plain APDU, wraps it with SM, sends to card, unwraps response.
func (s *smPureGo) Transmit(apdu []byte) (data []byte, sw uint16, err error) {
	if len(apdu) < 4 {
		return nil, 0, fmt.Errorf("APDU too short: %d bytes", len(apdu))
	}

	// Log plain APDU
	logVerbose("[APDU] Header: %X, bytes: %d\n", apdu[:4], len(apdu))

	// Parse APDU components
	cla := apdu[0]
	ins := apdu[1]
	p1 := apdu[2]
	p2 := apdu[3]

	var cmdData []byte
	var le int

	// Parse Lc and data
	if len(apdu) > 4 {
		if len(apdu) == 5 {
			// Case 2: CLA INS P1 P2 Le
			le = int(apdu[4])
			if le == 0 {
				le = 256
			}
		} else {
			// Case 3 or 4: CLA INS P1 P2 Lc [Data] [Le]
			lc := int(apdu[4])
			if len(apdu) >= 5+lc {
				cmdData = apdu[5 : 5+lc]

				// Check for Le
				if len(apdu) == 5+lc+1 {
					le = int(apdu[5+lc])
					if le == 0 {
						le = 256
					}
				}
			}
		}
	}

	// Wrap command with Secure Messaging
	wrappedCmd, err := s.smCtx.WrapCommand(cla, ins, p1, p2, cmdData, le)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to wrap command: %w", err)
	}

	// Transmit wrapped command to card
	response, err := s.card.Transmit(wrappedCmd)
	if err != nil {
		return nil, 0, fmt.Errorf("card transmit failed: %w", err)
	}

	// Unwrap response
	plainData, sw, err := s.smCtx.UnwrapResponse(response)
	if err != nil {
		logVerbose("[APDU] Unwrap error: %v\n", err)
		return nil, 0, fmt.Errorf("failed to unwrap response: %w", err)
	}

	logVerbose("[APDU] Response bytes: %d, sw=%04X\n", len(plainData), sw)

	return plainData, sw, nil
}
