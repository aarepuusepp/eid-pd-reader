package esteid

import (
	"fmt"
	"strings"
	"time"

	"github.com/ebfe/scard"
)

// CardResult represents a card read from a specific reader.
type CardResult struct {
	Reader       string        `json:"reader"`
	Duration     string        `json:"duration"`
	PersonalData *PersonalData `json:"data,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// perReaderTimeout limits how long we spend trying to read from a single reader.
const perReaderTimeout = 15 * time.Second

// ReadPersonalDataFromAllReaders reads from all available contact readers and returns results.
// NFC/contactless readers (PICC/SAM) are skipped — they require PACE via --nfc flag.
func ReadPersonalDataFromAllReaders(optionalFIDs map[string]string) ([]CardResult, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, fmt.Errorf("pcsc: establish context: %w", err)
	}
	defer ctx.Release()

	rdrs, err := ctx.ListReaders()
	if err != nil {
		return nil, fmt.Errorf("pcsc: list readers: %w", err)
	}
	if len(rdrs) == 0 {
		return nil, fmt.Errorf("pcsc: no readers found")
	}

	results := make([]CardResult, 0, len(rdrs))

	for _, readerName := range rdrs {
		// Skip NFC/contactless readers — they need PACE and the --nfc flag.
		if strings.Contains(readerName, "PICC") || strings.Contains(readerName, "SAM") {
			logVerbose("Skipping NFC reader in contact mode: %s", readerName)
			continue
		}

		result := readFromReaderWithTimeout(ctx, readerName, optionalFIDs)
		results = append(results, result)
	}

	return results, nil
}

// readFromReaderWithTimeout connects to a single reader, reads personal data, and
// enforces a timeout so that a stuck reader cannot block the entire operation.
func readFromReaderWithTimeout(ctx *scard.Context, readerName string, optionalFIDs map[string]string) CardResult {
	type readResult struct {
		pd  *PersonalData
		err error
	}

	ch := make(chan readResult, 1)
	start := time.Now()

	go func() {
		pd, err := connectAndRead(ctx, readerName, optionalFIDs)
		ch <- readResult{pd: pd, err: err}
	}()

	select {
	case res := <-ch:
		result := CardResult{Reader: readerName, Duration: FormatDuration(time.Since(start))}
		if res.err != nil {
			result.Error = res.err.Error()
		} else {
			result.PersonalData = res.pd
		}
		return result

	case <-time.After(perReaderTimeout):
		logVerbose("Timeout reading from reader: %s", readerName)
		return CardResult{
			Reader:   readerName,
			Duration: FormatDuration(time.Since(start)),
			Error:    fmt.Sprintf("timeout after %s", perReaderTimeout),
		}
	}
}

// FormatDuration formats a duration as human-readable "Xms" or "Xµs".
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

// connectAndRead handles the connect → read → disconnect sequence for one reader.
func connectAndRead(ctx *scard.Context, readerName string, optionalFIDs map[string]string) (*PersonalData, error) {
	card, err := ctx.Connect(readerName, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "SCARD_E_NO_SMARTCARD") ||
			strings.Contains(errMsg, "SCARD_W_REMOVED_CARD") ||
			strings.Contains(errMsg, "SCARD_W_UNRESPONSIVE_CARD") ||
			strings.Contains(errMsg, "error(8010000C)") ||
			strings.Contains(errMsg, "error(80100069)") {
			return nil, fmt.Errorf("no card present")
		}
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	defer card.Disconnect(scard.LeaveCard)

	return readPersonalDataFromCard(card, readerName, optionalFIDs)
}

// readPersonalDataFromCard reads from a specific card (extracted for reuse)
func readPersonalDataFromCard(card *scard.Card, readerName string, optionalFIDs map[string]string) (*PersonalData, error) {
	logVerbose("Reading from: %s", readerName)

	// Ensure card is ready before reading
	if err := ensureCardReady(card); err != nil {
		logVerbose("Card readiness check failed (will proceed anyway): %v", err)
	}

	// Detect card type from ATR
	status, err := card.Status()
	if err == nil {
		cardType := detectCardType(status.Atr)
		logVerbose("Card detected: %s (ATR: % X)", cardType, status.Atr)

		// Use optimized strategy based on card type
		switch cardType {
		case CardTypeThales, CardTypeThalesTest:
			logVerbose("Using Thales-optimized strategy")
			pd, err := readPersonalDataThales(card, optionalFIDs)
			if err == nil && validatePersonalData(pd) {
				return pd, nil
			}
			if err == nil {
				logVerbose("Thales strategy returned corrupted data, trying fallbacks")
			} else {
				logVerbose("Thales strategy failed: %v, trying fallbacks", err)
			}

		case CardTypeIdemiaCosom8, CardTypeIdemiaCosmoX:
			logVerbose("Using IDEMIA-optimized strategy")
			pd, err := readPersonalDataIdemia(card, optionalFIDs, cardType)
			if err == nil && validatePersonalData(pd) {
				return pd, nil
			}
			if err == nil {
				logVerbose("IDEMIA strategy returned corrupted data, trying fallbacks")
			} else {
				logVerbose("IDEMIA strategy failed: %v, trying fallbacks", err)
			}
		}
	}

	// Fallback: try standard approach
	logVerbose("Trying standard approach")
	pd, err := readPersonalDataStandard(card, optionalFIDs)
	if err == nil && validatePersonalData(pd) {
		return pd, nil
	}

	// If standard approach failed or returned corrupted data, try all strategies
	logVerbose("Standard approach failed or returned corrupted data, trying all strategies")
	
	strategies := []struct {
		name string
		fn   func(*scard.Card, map[string]string) (*PersonalData, error)
	}{
		{"Thales", func(c *scard.Card, fids map[string]string) (*PersonalData, error) {
			return readPersonalDataThales(c, fids)
		}},
		{"IDEMIA Cosmo 8", func(c *scard.Card, fids map[string]string) (*PersonalData, error) {
			return readPersonalDataIdemia(c, fids, CardTypeIdemiaCosom8)
		}},
		{"IDEMIA Cosmo X", func(c *scard.Card, fids map[string]string) (*PersonalData, error) {
			return readPersonalDataIdemia(c, fids, CardTypeIdemiaCosmoX)
		}},
	}

	for _, strategy := range strategies {
		logVerbose("Trying %s strategy", strategy.name)
		pd, err := strategy.fn(card, optionalFIDs)
		if err == nil && validatePersonalData(pd) {
			logVerbose("%s strategy succeeded", strategy.name)
			return pd, nil
		}
		if err != nil {
			logVerbose("%s strategy failed: %v", strategy.name, err)
		} else {
			logVerbose("%s strategy returned corrupted data", strategy.name)
		}
	}

	return nil, fmt.Errorf("all read strategies failed")
}

