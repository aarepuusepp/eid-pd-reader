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
	fmt.Printf("[PACE] Starting PACE protocol with CAN: %s\n", can)

	// Use AES-256 + brainpoolP256r1 as default (will be overridden by CardAccess)
	smCtx, err := pace.PerformPACE(
		cardAdapter,
		secret,
		pace.OID_PACE_ECDH_IM_AES256, // Will be updated from CardAccess
		16,                           // brainpoolP256r1 - will be updated from CardAccess
	)

	if err != nil {
		return nil, fmt.Errorf("PACE establishment failed: %w", err)
	}

	fmt.Printf("[PACE] Success! K_enc=%X K_mac=%X SSC0=%X\n",
		smCtx.EncryptionKey, smCtx.MACKey, smCtx.SSC)

	// Wrap the SM context in our adapter
	return NewSMPureGo(card, smCtx), nil
}

// PACEAvailable_Pure returns true since pure Go PACE is always available.
func PACEAvailable_Pure() bool {
	return true
}
