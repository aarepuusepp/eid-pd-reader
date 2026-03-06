package nfc

import (
	"fmt"
	"strings"
	"time"

	"github.com/ebfe/scard"
)

var verboseMode bool

// SetVerbose enables or disables verbose debug logging for NFC operations.
func SetVerbose(enabled bool) {
	verboseMode = enabled
}

func logVerbose(format string, args ...interface{}) {
	if verboseMode {
		fmt.Printf(format, args...)
	}
}

// ConnectAny connects to the first available NFC/contactless reader.
// It waits for a card to be present and retries the power-on sequence
// if the reader fails to reset the card (common with ACR1252 readers).
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

	// Collect contactless readers
	var nfcReaders []string
	for _, r := range rdrs {
		if IsContactlessReader(r) {
			nfcReaders = append(nfcReaders, r)
		}
	}

	if len(nfcReaders) == 0 {
		ctx.Release()
		return nil, nil, "", fmt.Errorf("no NFC/contactless reader found (available: %v)", rdrs)
	}

	// Wait for a card to be present on any NFC reader, then connect.
	card, readerName, err := waitAndConnect(ctx, nfcReaders)
	if err != nil {
		ctx.Release()
		return nil, nil, "", err
	}

	return ctx, card, readerName, nil
}

// waitAndConnect waits for a card to appear on one of the NFC readers,
// then attempts to connect with retries for power-on failures.
func waitAndConnect(ctx *scard.Context, readers []string) (*scard.Card, string, error) {
	// Build reader states for status monitoring
	states := make([]scard.ReaderState, len(readers))
	for i, r := range readers {
		states[i] = scard.ReaderState{
			Reader:       r,
			CurrentState: scard.StateUnaware,
		}
	}

	// Check current state first — a card may already be present
	err := ctx.GetStatusChange(states, time.Duration(0))
	if err != nil {
		logVerbose("[NFC] GetStatusChange failed: %v, falling back to direct connect\n", err)
		return directConnect(ctx, readers)
	}

	// If a card is already present on any reader, try to connect
	for i, st := range states {
		if st.EventState&scard.StatePresent != 0 {
			logVerbose("[NFC] Card already present on: %s\n", readers[i])
			card, err := connectSingle(ctx, readers[i])
			if err == nil {
				return card, readers[i], nil
			}
			logVerbose("[NFC] Connect failed on %s: %v\n", readers[i], err)
		}
		// Update current state for the wait loop
		states[i].CurrentState = st.EventState
	}

	// No card present yet — wait up to 5 seconds for a card to appear
	logVerbose("[NFC] No card present, waiting up to 5s for card...\n")
	err = ctx.GetStatusChange(states, 5*time.Second)
	if err != nil {
		return nil, "", fmt.Errorf("no NFC reader available or no card present (timeout waiting for card)")
	}

	for i, st := range states {
		if st.EventState&scard.StatePresent != 0 {
			logVerbose("[NFC] Card detected on: %s\n", readers[i])
			card, err := connectSingle(ctx, readers[i])
			if err == nil {
				return card, readers[i], nil
			}
			logVerbose("[NFC] Connect failed on %s: %v\n", readers[i], err)
		}
	}

	return nil, "", fmt.Errorf("no NFC reader available or no card present")
}

// connectSingle attempts a single SCardConnect after a stabilization delay.
// Each failed SCardConnect triggers 3-4 Event 610 entries in the Windows
// Event Viewer (the SCard service internally retries POWER commands).
// To minimize this noise, we make exactly ONE attempt per call and rely
// on the higher-level read retry loop for recovery.
func connectSingle(ctx *scard.Context, reader string) (*scard.Card, error) {
	time.Sleep(500 * time.Millisecond)

	card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		logVerbose("[NFC] SCardConnect to %s failed: %v\n", reader, err)
		return nil, err
	}
	return card, nil
}

// directConnect is the fallback when GetStatusChange is not supported.
func directConnect(ctx *scard.Context, readers []string) (*scard.Card, string, error) {
	for _, r := range readers {
		card, err := connectSingle(ctx, r)
		if err == nil {
			return card, r, nil
		}
	}
	return nil, "", fmt.Errorf("no NFC reader available or no card present")
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
	if IsContactlessReader(readerName) {
		return true
	}
	return false
}
