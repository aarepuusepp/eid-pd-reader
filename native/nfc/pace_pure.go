package nfc

import (
	"fmt"

	"github.com/aarepuusepp/libopenpace-go/pace"
	"github.com/ebfe/scard"
)

// StartPACE establishes a PACE session using pure Go implementation.
//
// This is the pure Go implementation that doesn't require OpenPACE or CGO.
// It returns an SM interface compatible with the existing eid-pd-reader code.
//
// Parameters:
//   - card: PC/SC smart card
//   - can: Card Access Number (6 digits)
//
// Returns an SM interface for protected communication.
func StartPACE(card *scard.Card, can string) (SM, error) {
	if len(can) != 6 {
		return nil, fmt.Errorf("CAN must be 6 digits, got %d", len(can))
	}

	// Create card adapter for libopenpace-go
	cardAdapter := NewCardAdapter(card)

	// Prepare PACE secret (CAN)
	secret := &pace.PACESecret{
		Type:  pace.SecretTypeCAN,
		Value: []byte(can),
	}

	// PerformPACE will read EF.CardAccess and determine the correct
	// protocol (AES-128/192/256) and curve (P-256, brainpool, etc.)
	// We pass default values here, but PerformPACE will override them
	// based on what the card advertises in CardAccess
	logVerbose("[PACE] Starting PACE protocol\n")

	// Defaults for IDEMIA cards (ECDH-GM-AES256 + secp256r1).
	// PerformPACE reads EF.CardAccess and overrides with the actual values.
	smCtx, err := pace.PerformPACE(
		cardAdapter,
		secret,
		pace.OID_PACE_ECDH_GM_AES256,
		12, // secp256r1/NIST P-256 (BSI TR-03110 ParamID 12), overridden from CardAccess
	)

	if err != nil {
		return nil, fmt.Errorf("PACE establishment failed: %w", err)
	}

	logVerbose("[PACE] Secure messaging established\n")

	// Wrap the SM context in our adapter
	return NewSMPureGo(card, smCtx), nil
}

// PACEAvailable_Pure returns true since pure Go PACE is always available.
func PACEAvailable_Pure() bool {
	return true
}
