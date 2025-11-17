package esteid

import (
	"fmt"
	"strings"

	"github.com/ebfe/scard"
)

// CardResult represents a card read from a specific reader
type CardResult struct {
	Reader       string        `json:"reader"`
	PersonalData *PersonalData `json:"data,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// ReadPersonalDataFromAllReaders reads from all available readers and returns results
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

	// Try to read from each reader
	for _, readerName := range rdrs {
		result := CardResult{Reader: readerName}
		
		card, err := ctx.Connect(readerName, scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			errMsg := err.Error()
			// Check for "no card" errors
			if strings.Contains(errMsg, "SCARD_E_NO_SMARTCARD") ||
				strings.Contains(errMsg, "SCARD_W_REMOVED_CARD") ||
				strings.Contains(errMsg, "error(8010000C)") {
				result.Error = "no card present"
			} else {
				result.Error = fmt.Sprintf("connection failed: %v", err)
			}
			results = append(results, result)
			continue
		}

		// Try to read personal data from this card
		pd, err := readPersonalDataFromCard(card, readerName, optionalFIDs)
		card.Disconnect(scard.LeaveCard)
		
		if err != nil {
			result.Error = err.Error()
		} else {
			result.PersonalData = pd
		}
		results = append(results, result)
	}

	return results, nil
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

