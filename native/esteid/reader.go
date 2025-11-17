package esteid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ebfe/scard"

	"ee.aarepuusepp.eid-pd-reader/native/apdu"
)

var verboseMode bool

// SetVerbose enables or disables verbose logging
func SetVerbose(enabled bool) {
	verboseMode = enabled
}

func logVerbose(format string, args ...interface{}) {
	if verboseMode {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Default file IDs (FID) for EstEID Personal Data (transparent EFs)
var (
	FID_DF_PERSONAL  = [2]byte{0x50, 0x00} // DF '5000'
	FID_SURNAME      = [2]byte{0x50, 0x01} // EF '5001' - PD1
	FID_GIVENNAME    = [2]byte{0x50, 0x02} // EF '5002' - PD2
	FID_SEX          = [2]byte{0x50, 0x03} // EF '5003' - PD3
	FID_CITIZENSHIP  = [2]byte{0x50, 0x04} // EF '5004' - PD4
	FID_BIRTHDATE    = [2]byte{0x50, 0x05} // EF '5005' - PD5
	FID_PERSONALCODE = [2]byte{0x50, 0x06} // EF '5006' - PD6
	FID_DOCUMENTNR   = [2]byte{0x50, 0x07} // EF '5007' - PD7
	FID_EXPIRYDATE   = [2]byte{0x50, 0x08} // EF '5008' - PD8
	FID_ISSUEDATE    = [2]byte{0x50, 0x09} // EF '5009' - PD9
)

// PersonalData holds the fields we export
type PersonalData struct {
	Surname      string `json:"surname"`
	GivenName    string `json:"givenName"`
	Sex          string `json:"sex,omitempty"`
	Citizenship  string `json:"citizenship,omitempty"`
	BirthDate    string `json:"birthDate,omitempty"`
	PersonalCode string `json:"personalCode"`
	DocumentNr   string `json:"documentNr,omitempty"`
	ExpiryDate   string `json:"expiryDate,omitempty"`
	IssueDate    string `json:"issueDate,omitempty"`
	Source       string `json:"source"` // "chip:EF500x"
}

// Known ATRs for Estonian ID cards
var (
	// IDEMIA Cosmo 8.1/8.2
	ATR_IDEMIA_COSMO8 = []byte{0x3B, 0xDB, 0x96, 0x00, 0x80, 0xB1, 0xFE, 0x45, 0x1F, 0x83, 0x00, 0x12, 0x23, 0x3F, 0x53, 0x65, 0x49, 0x44, 0x0F, 0x90, 0x00, 0xF1}
	// IDEMIA Cosmo X
	ATR_IDEMIA_COSMOX = []byte{0x3B, 0xDC, 0x96, 0x00, 0x80, 0xB1, 0xFE, 0x45, 0x1F, 0x83, 0x00, 0x12, 0x23, 0x3F, 0x54, 0x65, 0x49, 0x44, 0x32, 0x0F, 0x90, 0x00, 0xC3}
	// Thales (production)
	ATR_THALES = []byte{0x3B, 0xFF, 0x96, 0x00, 0x00, 0x80, 0x31, 0xFE, 0x43, 0x80, 0x31, 0xB8, 0x53, 0x65, 0x49, 0x44, 0x64, 0xB0, 0x85, 0x05, 0x10, 0x12, 0x23, 0x3F, 0x1D}
	// Thales test cards / NFC cards (same ATR as some IDEMIA cards)
	ATR_THALES_TEST = []byte{0x3B, 0x8F, 0x80, 0x01, 0x80, 0x31, 0xD8, 0x53, 0x65, 0x49, 0x44, 0x64, 0xB0, 0x85, 0x05, 0x10, 0x12, 0x23, 0x3F, 0x16}
)

// CardType represents the detected card manufacturer
type CardType int

const (
	CardTypeUnknown CardType = iota
	CardTypeIdemiaCosom8
	CardTypeIdemiaCosmoX
	CardTypeThales
	CardTypeThalesTest
)

func (ct CardType) String() string {
	switch ct {
	case CardTypeIdemiaCosom8:
		return "IDEMIA Cosmo 8.1/8.2"
	case CardTypeIdemiaCosmoX:
		return "IDEMIA Cosmo X"
	case CardTypeThales:
		return "Thales"
	case CardTypeThalesTest:
		return "Thales (test/NFC)"
	default:
		return "Unknown"
	}
}

// detectCardType identifies the card type based on ATR
func detectCardType(atr []byte) CardType {
	// Check specific ATRs first
	if bytes.Equal(atr, ATR_IDEMIA_COSMO8) {
		return CardTypeIdemiaCosom8
	}
	if bytes.Equal(atr, ATR_IDEMIA_COSMOX) {
		return CardTypeIdemiaCosmoX
	}
	if bytes.Equal(atr, ATR_THALES) {
		return CardTypeThales
	}
	// This ATR is shared by Thales test cards and some IDEMIA cards
	// Default to Thales test card (more common with NFC readers)
	if bytes.Equal(atr, ATR_THALES_TEST) {
		return CardTypeThalesTest
	}
	return CardTypeUnknown
}

// ensureCardReady verifies the card is ready to accept commands
// This prevents reading corrupted data when card was just inserted
func ensureCardReady(card *scard.Card) error {
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logVerbose("Card readiness check attempt %d/%d", attempt+1, maxRetries)
			time.Sleep(retryDelay)
		}

		// Try to get card status
		status, err := card.Status()
		if err != nil {
			logVerbose("Card status check failed: %v", err)
			continue
		}

		// Check if card is responsive
		if len(status.Atr) == 0 {
			logVerbose("Card ATR is empty, card may not be ready")
			continue
		}

		// Try a simple SELECT MF command to verify card is responsive
		// Some cards are picky about P1/P2 values
		// Try multiple variants to find what works
		selectVariants := [][]byte{
			{0x00, 0xA4, 0x00, 0x0C, 0x02, 0x3F, 0x00}, // P2=0x0C (FCI, first/only)
			{0x00, 0xA4, 0x00, 0x00, 0x02, 0x3F, 0x00}, // P2=0x00 (return FCI)
			{0x00, 0xA4, 0x01, 0x0C, 0x02, 0x3F, 0x00}, // P1=0x01 (select child DF)
		}

		cardReady := false
		for _, variant := range selectVariants {
			resp, err := card.Transmit(variant)
			if err != nil {
				continue
			}
			if len(resp) >= 2 {
				sw1, sw2 := resp[len(resp)-2], resp[len(resp)-1]
				if sw1 == 0x90 && sw2 == 0x00 {
					logVerbose("Card is ready (attempt %d)", attempt+1)
					cardReady = true
					break
				}
				logVerbose("SELECT MF variant returned SW=%02X%02X", sw1, sw2)
			}
		}

		if cardReady {
			return nil
		}

		logVerbose("All SELECT MF variants failed on attempt %d", attempt+1)
	}

	return fmt.Errorf("card not responding after %d attempts", maxRetries)
}

// ReadPersonalData reads PD EFs via PC/SC from the first available EstEID card.
// Overrides optionalFIDs can provide custom FIDs in hex like ["5000","5001",...].
func ReadPersonalData(optionalFIDs map[string]string) (*PersonalData, error) {
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

	var card *scard.Card
	var lastErr error
	var connectedReader string

	// Try to connect to a reader with card
	// Skip NFC readers (PICC/SAM) in contact mode - they need PACE
	for _, r := range rdrs {
		// Skip NFC readers in contact mode (they need --nfc flag with PACE)
		if strings.Contains(r, "PICC") || strings.Contains(r, "SAM") {
			logVerbose("Skipping NFC reader in contact mode: %s", r)
			continue
		}

		c, err := ctx.Connect(r, scard.ShareShared, scard.ProtocolAny)
		if err == nil {
			card = c
			connectedReader = r
			break
		}
		lastErr = err
	}
	if card == nil {
		if lastErr != nil {
			errMsg := lastErr.Error()
			// Check for common "no card" error codes
			if strings.Contains(errMsg, "SCARD_E_NO_SMARTCARD") ||
				strings.Contains(errMsg, "SCARD_W_REMOVED_CARD") ||
				strings.Contains(errMsg, "SCARD_W_UNRESPONSIVE_CARD") ||
				strings.Contains(errMsg, "error(8010000C)") || // SCARD_E_NO_SMARTCARD
				strings.Contains(errMsg, "error(80100069)") { // SCARD_W_UNRESPONSIVE_CARD
				return nil, fmt.Errorf("CARD_NOT_FOUND")
			}
			return nil, fmt.Errorf("pcsc: could not connect to any reader (last error: %v)", lastErr)
		}
		return nil, fmt.Errorf("CARD_NOT_FOUND")
	}
	defer card.Disconnect(scard.LeaveCard)

	logVerbose("Connected to reader: %s", connectedReader)

	// Ensure card is ready before reading
	// This prevents reading corrupted data when card was just inserted
	// If readiness check fails, we'll proceed anyway - some cards don't respond to SELECT MF
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
		default:
			logVerbose("Unknown card type, trying all strategies")
		}
	}

	// Try multiple strategies in order of likelihood
	// Strategy 1: GET DATA approach (some cards support this)
	logVerbose("\n=== Trying Strategy 1: GET DATA ===")
	pd, err := readPersonalDataViaGetData(card)
	if err == nil {
		if validatePersonalData(pd) {
			logVerbose("✓ Strategy 1 succeeded")
			return pd, nil
		}
		logVerbose("✗ Strategy 1 returned corrupted data")
	} else {
		logVerbose("✗ Strategy 1 failed: %v", err)
	}

	// Strategy 2: Standard file-based approach with DF
	logVerbose("\n=== Trying Strategy 2: Standard (MF->DF->EF) ===")
	pd, err = readPersonalDataStandard(card, optionalFIDs)
	if err == nil {
		if validatePersonalData(pd) {
			logVerbose("✓ Strategy 2 succeeded")
			return pd, nil
		}
		logVerbose("✗ Strategy 2 returned corrupted data")
	} else {
		logVerbose("✗ Strategy 2 failed: %v", err)
	}

	// Strategy 3: Direct EF selection without DF (some Thales cards)
	logVerbose("\n=== Trying Strategy 3: Direct EF (no DF) ===")
	pd, err = readPersonalDataDirectEF(card, optionalFIDs)
	if err == nil {
		if validatePersonalData(pd) {
			logVerbose("✓ Strategy 3 succeeded")
			return pd, nil
		}
		logVerbose("✗ Strategy 3 returned corrupted data")
	} else {
		logVerbose("✗ Strategy 3 failed: %v", err)
	}

	// Strategy 4: Path-based selection (full path from MF)
	logVerbose("\n=== Trying Strategy 4: Path-based ===")
	pd, err = readPersonalDataByPath(card, optionalFIDs)
	if err == nil {
		if validatePersonalData(pd) {
			logVerbose("✓ Strategy 4 succeeded")
			return pd, nil
		}
		logVerbose("✗ Strategy 4 returned corrupted data")
	} else {
		logVerbose("✗ Strategy 4 failed: %v", err)
	}

	// All strategies failed
	return nil, fmt.Errorf("all read strategies failed. Last error: %w", err)
}

// readPersonalDataThales uses Thales-specific command sequence
func readPersonalDataThales(card *scard.Card, optionalFIDs map[string]string) (*PersonalData, error) {
	// Try Method 1: SELECT MF then try direct EF access (for test cards via NFC)
	logVerbose("Method 1: MF + direct EF access")
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{0x3F, 0x00}, Le: -1}
	resp, err := transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			logVerbose("MF selected, trying direct EF reads")
			// Try to read EF 5001 (surname) directly
			cmd = apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x50, P2: 0x01, Le: 0}
			resp, err = transmit(card, cmd)
			if err == nil {
				sw1, sw2 = apdu.StatusWord(resp)
				if sw1 == 0x90 || sw1 == 0x61 || sw1 == 0x6C {
					logVerbose("Direct READ BINARY worked! Trying full read...")
					// This might work - try the standard read
					return readAllPersonalDataFields(card, optionalFIDs, "chip:Thales-Test")
				}
			}
		}
	}

	// Try Method 2: SELECT EstEID AID directly (works for some test cards)
	logVerbose("Method 2: EstEID AID selection")
	esteidAID := []byte{0xA0, 0x00, 0x00, 0x00, 0x77, 0x01, 0x08, 0x00, 0x07, 0x00, 0x00, 0xFE, 0x00, 0x00, 0x01, 0x00}
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x04, P2: 0x00, Data: esteidAID, Le: -1}
	resp, err = transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			logVerbose("EstEID AID selected successfully")
			// Now try to select DF 5000
			cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{0x50, 0x00}, Le: -1}
			resp, err = transmit(card, cmd)
			if err == nil {
				sw1, sw2 = apdu.StatusWord(resp)
				if sw1 == 0x90 && sw2 == 0x00 {
					logVerbose("DF 5000 selected successfully")
					return readAllPersonalDataFields(card, optionalFIDs, "chip:Thales")
				}
			}
		}
	}

	// Try Method 3: SELECT PKCS-15 AID (traditional method for production cards)
	logVerbose("Method 3: PKCS-15 AID selection")
	pkcs15AID := []byte{0xA0, 0x00, 0x00, 0x00, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35}
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x04, P2: 0x00, Data: pkcs15AID, Le: -1}
	resp, err = transmit(card, cmd)
	if err != nil {
		return nil, err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	if sw1 != 0x90 || sw2 != 0x00 {
		return nil, fmt.Errorf("failed to select PKCS-15 AID: SW=%02X%02X", sw1, sw2)
	}

	// SELECT DF DFDD (Personal Data DF for Thales)
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x08, P2: 0x0C, Data: []byte{0xDF, 0xDD}, Le: -1}
	resp, err = transmit(card, cmd)
	if err != nil {
		return nil, err
	}
	sw1, sw2 = apdu.StatusWord(resp)
	if sw1 != 0x90 || sw2 != 0x00 {
		return nil, fmt.Errorf("failed to select DF DFDD: SW=%02X%02X", sw1, sw2)
	}

	return readAllPersonalDataFields(card, optionalFIDs, "chip:Thales")
}

// readPersonalDataIdemia uses IDEMIA-specific command sequence
func readPersonalDataIdemia(card *scard.Card, optionalFIDs map[string]string, cardType CardType) (*PersonalData, error) {
	// SELECT MF then DF 5000 using path-based selection
	// 00 A4 09 0C 04 3F00 5000
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x09, P2: 0x0C, Data: []byte{0x3F, 0x00, 0x50, 0x00}, Le: -1}
	resp, err := transmit(card, cmd)
	if err != nil {
		return nil, err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	if sw1 != 0x90 || sw2 != 0x00 {
		return nil, fmt.Errorf("failed to select MF/DF: SW=%02X%02X", sw1, sw2)
	}

	source := "chip:IDEMIA"
	if cardType == CardTypeIdemiaCosmoX {
		source = "chip:IDEMIA-CosmoX"
	}

	return readAllPersonalDataFields(card, optionalFIDs, source)
}

// readPersonalDataStandard uses the standard approach: SELECT MF -> SELECT DF -> SELECT EF -> READ
func readPersonalDataStandard(card *scard.Card, optionalFIDs map[string]string) (*PersonalData, error) {
	// SELECT MF 3F00
	if err := selectByFID(card, [2]byte{0x3F, 0x00}); err != nil {
		return nil, fmt.Errorf("select MF: %w", err)
	}

	// DF Personal (5000) — allow override
	pdfid := FID_DF_PERSONAL
	if s, ok := optionalFIDs["df_personal"]; ok {
		if fid, e := parseFID(s); e == nil {
			pdfid = fid
		}
	}

	// Try to select DF, but continue even if it fails
	dfSelected := false
	if err := selectDF(card, pdfid); err == nil {
		dfSelected = true
	}

	if !dfSelected {
		return nil, fmt.Errorf("DF selection failed")
	}

	// Read all fields
	return readAllPersonalDataFields(card, optionalFIDs, "chip:EF500x")
}

// readPersonalDataDirectEF tries to read EFs directly without DF selection
// Some Thales cards reject DF selection but allow direct EF access
func readPersonalDataDirectEF(card *scard.Card, optionalFIDs map[string]string) (*PersonalData, error) {
	// Thales cards require selecting PKCS-15 AID first, then DF DFDD
	logVerbose("Trying Thales-specific sequence: AID -> DFDD -> EFs")

	// SELECT PKCS-15 AID: A000000063504B43532D3135
	pkcs15AID := []byte{0xA0, 0x00, 0x00, 0x00, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35}
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x04, P2: 0x00, Data: pkcs15AID, Le: -1}
	resp, err := transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			logVerbose("✓ PKCS-15 AID selected")

			// SELECT DF DFDD (Personal Data DF for Thales)
			cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x08, P2: 0x0C, Data: []byte{0xDF, 0xDD}, Le: -1}
			resp, err = transmit(card, cmd)
			if err == nil {
				sw1, sw2 = apdu.StatusWord(resp)
				if sw1 == 0x90 && sw2 == 0x00 {
					logVerbose("✓ DF DFDD selected, reading EFs...")
					return readAllPersonalDataFields(card, optionalFIDs, "chip:Thales")
				}
			}
		}
	}

	// Fallback: Try without AID selection
	logVerbose("Thales sequence failed, trying direct EF access")
	return readAllPersonalDataFields(card, optionalFIDs, "chip:DirectEF")
}

// readPersonalDataByPath uses full path selection (3F00/5000/5001 etc)
func readPersonalDataByPath(card *scard.Card, optionalFIDs map[string]string) (*PersonalData, error) {
	// Some cards prefer path-based selection
	// SELECT by path: P1=0x08 (path from MF), P1=0x09 (path from current DF)

	// Get FIDs
	surnameFID := FID_SURNAME
	if s, ok := optionalFIDs["ef_surname"]; ok {
		if fid, e := parseFID(s); e == nil {
			surnameFID = fid
		}
	}

	// Try reading surname with path-based selection
	surname, err := readEFByPath(card, surnameFID, 256)
	if err != nil {
		return nil, err
	}

	givenFID := FID_GIVENNAME
	if s, ok := optionalFIDs["ef_given"]; ok {
		if fid, e := parseFID(s); e == nil {
			givenFID = fid
		}
	}
	given, err := readEFByPath(card, givenFID, 256)
	if err != nil {
		return nil, err
	}

	codeFID := FID_PERSONALCODE
	if s, ok := optionalFIDs["ef_code"]; ok {
		if fid, e := parseFID(s); e == nil {
			codeFID = fid
		}
	}
	code, err := readEFByPath(card, codeFID, 256)
	if err != nil {
		return nil, err
	}

	// Read optional fields
	sex, _ := readEFByPath(card, FID_SEX, 256)
	citizenship, _ := readEFByPath(card, FID_CITIZENSHIP, 256)
	birthDate, _ := readEFByPath(card, FID_BIRTHDATE, 256)
	documentNr, _ := readEFByPath(card, FID_DOCUMENTNR, 256)
	expiryDate, _ := readEFByPath(card, FID_EXPIRYDATE, 256)
	issueDate, _ := readEFByPath(card, FID_ISSUEDATE, 256)

	return &PersonalData{
		Surname:      sanitizeText(surname),
		GivenName:    sanitizeText(given),
		Sex:          sanitizeText(sex),
		Citizenship:  sanitizeText(citizenship),
		BirthDate:    sanitizeText(birthDate),
		PersonalCode: sanitizeText(code),
		DocumentNr:   sanitizeText(documentNr),
		ExpiryDate:   sanitizeText(expiryDate),
		IssueDate:    sanitizeText(issueDate),
		Source:       "chip:PathBased",
	}, nil
}

// readAllPersonalDataFields reads all PD fields using standard EF selection
func readAllPersonalDataFields(card *scard.Card, optionalFIDs map[string]string, source string) (*PersonalData, error) {
	// Get FIDs with overrides
	surnameFID := FID_SURNAME
	if s, ok := optionalFIDs["ef_surname"]; ok {
		if fid, e := parseFID(s); e == nil {
			surnameFID = fid
		}
	}
	givenFID := FID_GIVENNAME
	if s, ok := optionalFIDs["ef_given"]; ok {
		if fid, e := parseFID(s); e == nil {
			givenFID = fid
		}
	}
	codeFID := FID_PERSONALCODE
	if s, ok := optionalFIDs["ef_code"]; ok {
		if fid, e := parseFID(s); e == nil {
			codeFID = fid
		}
	}

	// Read required fields
	logVerbose("\n--- Reading Surname (FID=%02X%02X) ---", surnameFID[0], surnameFID[1])
	surname, err := readTransparentEF(card, surnameFID, 256)
	if err != nil {
		return nil, fmt.Errorf("read EF %02X%02X: %w", surnameFID[0], surnameFID[1], err)
	}
	logVerbose("Raw surname data: % X", surname)
	logVerbose("Surname text: %q", string(surname))

	// Validate surname - if it contains binary/certificate data, it's wrong
	surnameText := sanitizeText(surname)
	if containsBinaryData(surnameText) {
		logVerbose("⚠ Surname contains binary data, marking as corrupted")
		surnameText = ""
	}
	logVerbose("Final surname: %q", surnameText)

	logVerbose("\n--- Reading Given Name (FID=%02X%02X) ---", givenFID[0], givenFID[1])
	given, err := readTransparentEF(card, givenFID, 256)
	if err != nil {
		return nil, fmt.Errorf("read EF %02X%02X: %w", givenFID[0], givenFID[1], err)
	}
	logVerbose("Raw given name data: % X", given)
	logVerbose("Given name text: %q", string(given))

	logVerbose("\n--- Reading Personal Code (FID=%02X%02X) ---", codeFID[0], codeFID[1])
	code, err := readTransparentEF(card, codeFID, 256)
	if err != nil {
		return nil, fmt.Errorf("read EF %02X%02X: %w", codeFID[0], codeFID[1], err)
	}
	logVerbose("Raw personal code data: % X", code)
	logVerbose("Personal code text: %q", string(code))

	// Read optional fields (don't fail if they're not present)
	sex, _ := readTransparentEF(card, FID_SEX, 256)
	citizenship, _ := readTransparentEF(card, FID_CITIZENSHIP, 256)
	birthDate, _ := readTransparentEF(card, FID_BIRTHDATE, 256)
	documentNr, _ := readTransparentEF(card, FID_DOCUMENTNR, 256)
	expiryDate, _ := readTransparentEF(card, FID_EXPIRYDATE, 256)
	issueDate, _ := readTransparentEF(card, FID_ISSUEDATE, 256)

	return &PersonalData{
		Surname:      surnameText,
		GivenName:    sanitizeText(given),
		Sex:          sanitizeText(sex),
		Citizenship:  sanitizeText(citizenship),
		BirthDate:    sanitizeText(birthDate),
		PersonalCode: sanitizeText(code),
		DocumentNr:   sanitizeText(documentNr),
		ExpiryDate:   sanitizeText(expiryDate),
		IssueDate:    sanitizeText(issueDate),
		Source:       source,
	}, nil
}

// containsBinaryData checks if text contains binary/certificate data
func containsBinaryData(s string) bool {
	// Check for common certificate/key markers
	if strings.Contains(s, "key") || strings.Contains(s, "certificate") {
		return true
	}
	// Check for high percentage of non-printable characters
	nonPrintable := 0
	for _, r := range s {
		if r < 32 || r > 126 {
			nonPrintable++
		}
	}
	if len(s) > 0 && float64(nonPrintable)/float64(len(s)) > 0.3 {
		return true
	}
	return false
}

func parseFID(s string) ([2]byte, error) {
	s = strings.TrimPrefix(strings.ToLower(s), "0x")
	if len(s) != 4 {
		return [2]byte{}, fmt.Errorf("fid must be 4 hex chars")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return [2]byte{}, err
	}
	return [2]byte{b[0], b[1]}, nil
}

func selectByFID(card *scard.Card, fid [2]byte) error {
	// 00 A4 00 0C 02 FID
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err := transmit(card, cmd)
	if err != nil {
		return err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
}

func selectDF(card *scard.Card, fid [2]byte) error {
	// Try multiple SELECT methods for compatibility with different card profiles
	// Method 1: Select by FID (P1=0x00) - works for most Estonian ID cards
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err := transmit(card, cmd)
	if err != nil {
		return err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	if sw1 == 0x90 && sw2 == 0x00 {
		return nil // Success
	}

	// Method 2: Select child DF (P1=0x01) - fallback for some profiles
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x01, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err = transmit(card, cmd)
	if err != nil {
		return err
	}
	sw1, sw2 = apdu.StatusWord(resp)
	if sw1 == 0x90 && sw2 == 0x00 {
		return nil // Success
	}

	// Method 3: Select by path from MF (P1=0x08) - for some card profiles
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x08, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err = transmit(card, cmd)
	if err != nil {
		return err
	}
	sw1, sw2 = apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
}

func selectEF(card *scard.Card, fid [2]byte) error {
	// Try multiple P1 values for EF selection
	// Method 1: P1=0x02 (select EF under current DF)
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x02, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err := transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			return nil
		}
	}

	// Method 2: P1=0x00 (select by file ID from MF)
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err = transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			return nil
		}
	}

	// Method 3: Try with P2=0x04 (return FCI)
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x02, P2: 0x04, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err = transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			return nil
		}
	}

	return fmt.Errorf("all EF selection methods failed")
}

func readTransparentEF(card *scard.Card, fid [2]byte, max int) ([]byte, error) {
	if err := selectEF(card, fid); err != nil {
		return nil, err
	}
	// Read up to max bytes using READ BINARY (00 B0 offHi offLo Le). Most PD fields are short.
	cmd := apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: max}
	resp, err := transmit(card, cmd)
	if err != nil {
		return nil, err
	}
	sw1, sw2 := apdu.StatusWord(resp)

	// Handle SW=6CXX (Wrong Le) - card tells us the correct length
	if sw1 == 0x6C {
		// SW2 contains the correct length to request
		correctLen := int(sw2)
		cmd = apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: correctLen}
		resp, err = transmit(card, cmd)
		if err != nil {
			return nil, err
		}
		sw1, sw2 = apdu.StatusWord(resp)
	}

	if err := apdu.CheckOK(sw1, sw2); err != nil {
		return nil, err
	}
	return bytes.TrimRight(apdu.DataPart(resp), "\x00\x20\n\r"), nil
}

func transmit(card *scard.Card, cmd apdu.APDU) ([]byte, error) {
	cmdBytes := cmd.Bytes()
	logVerbose(">>> APDU: % X", cmdBytes)

	resp, err := card.Transmit(cmdBytes)
	if err != nil {
		logVerbose("<<< ERROR: %v", err)
		return nil, fmt.Errorf("pcsc: transmit % X: %w", cmdBytes, err)
	}

	if len(resp) >= 2 {
		sw1, sw2 := resp[len(resp)-2], resp[len(resp)-1]
		logVerbose("<<< Response: % X (SW=%02X%02X, %d bytes data)", resp, sw1, sw2, len(resp)-2)
	} else {
		logVerbose("<<< Response: % X", resp)
	}

	return resp, nil
}

// readEFByPath reads an EF using path-based selection (full path from MF)
func readEFByPath(card *scard.Card, fid [2]byte, max int) ([]byte, error) {
	// Try different path-based selection methods

	// Method 1: SELECT by path from MF with P2=0x00 (first occurrence)
	// Path: 3F00 (MF) / 5000 (DF) / 50XX (EF)
	path := []byte{0x3F, 0x00, 0x50, 0x00, fid[0], fid[1]}
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x08, P2: 0x00, Data: path, Le: -1}
	resp, err := transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			// Success, now read
			return readBinaryWithRetry(card, max)
		}
	}

	// Method 2: Try with P2=0x0C (return FCP)
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x08, P2: 0x0C, Data: path, Le: -1}
	resp, err = transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			// Success, now read
			return readBinaryWithRetry(card, max)
		}
	}

	// Method 3: Try shorter path (just DF + EF, assuming MF already selected)
	shortPath := []byte{0x50, 0x00, fid[0], fid[1]}
	cmd = apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x08, P2: 0x0C, Data: shortPath, Le: -1}
	resp, err = transmit(card, cmd)
	if err == nil {
		sw1, sw2 := apdu.StatusWord(resp)
		if sw1 == 0x90 && sw2 == 0x00 {
			// Success, now read
			return readBinaryWithRetry(card, max)
		}
	}

	return nil, fmt.Errorf("all path-based selection methods failed")
}

// readBinaryWithRetry reads binary data and handles SW=6CXX (Wrong Le)
func readBinaryWithRetry(card *scard.Card, max int) ([]byte, error) {
	readCmd := apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: max}
	resp, err := transmit(card, readCmd)
	if err != nil {
		return nil, err
	}
	sw1, sw2 := apdu.StatusWord(resp)

	// Handle SW=6CXX (Wrong Le)
	if sw1 == 0x6C {
		correctLen := int(sw2)
		readCmd = apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: correctLen}
		resp, err = transmit(card, readCmd)
		if err != nil {
			return nil, err
		}
		sw1, sw2 = apdu.StatusWord(resp)
	}

	if err := apdu.CheckOK(sw1, sw2); err != nil {
		return nil, err
	}
	return bytes.TrimRight(apdu.DataPart(resp), "\x00\x20\n\r"), nil
}

func sanitizeText(b []byte) string {
	// EstEID PD tends to be ASCII/UTF-8. Remove trailing control chars.
	out := string(bytes.TrimSpace(b))
	// Collapse internal NULs (some cards may pad with 0x00)
	out = strings.ReplaceAll(out, "\x00", "")
	return out
}

// isValidUTF8Text checks if the text contains valid UTF-8 and no replacement characters
func isValidUTF8Text(s string) bool {
	// Check for UTF-8 replacement character (�) which indicates corrupted encoding
	if strings.Contains(s, "\uFFFD") {
		return false
	}
	// Check for excessive control characters (except space and common whitespace)
	controlCount := 0
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			controlCount++
		}
	}
	// If more than 20% of characters are control characters, data is likely corrupted
	if len(s) > 0 && controlCount > len(s)/5 {
		return false
	}
	return true
}

// validatePersonalData checks if personal data contains valid text
func validatePersonalData(pd *PersonalData) bool {
	// Check critical fields for corruption
	if pd.Surname != "" && !isValidUTF8Text(pd.Surname) {
		logVerbose("Surname contains invalid UTF-8: %q", pd.Surname)
		return false
	}
	if pd.GivenName != "" && !isValidUTF8Text(pd.GivenName) {
		logVerbose("GivenName contains invalid UTF-8: %q", pd.GivenName)
		return false
	}
	if pd.PersonalCode != "" && !isValidUTF8Text(pd.PersonalCode) {
		logVerbose("PersonalCode contains invalid UTF-8: %q", pd.PersonalCode)
		return false
	}
	return true
}

// readPersonalDataViaGetData attempts to read PD using GET DATA commands
// This works with some Thales cards that don't support file navigation
func readPersonalDataViaGetData(card *scard.Card) (*PersonalData, error) {
	// GET DATA tags for Estonian ID card personal data
	// These are BER-TLV tags commonly used in ID cards
	tags := map[string][2]byte{
		"surname":      {0x5F, 0x0E}, // Tag for surname
		"givenName":    {0x5F, 0x0D}, // Tag for given names
		"personalCode": {0x5F, 0x10}, // Tag for personal code
		"sex":          {0x5F, 0x35}, // Tag for sex
		"citizenship":  {0x5F, 0x2A}, // Tag for nationality
		"birthDate":    {0x5F, 0x2B}, // Tag for date of birth
		"documentNr":   {0x5A, 0x00}, // Tag for document number
		"expiryDate":   {0x5F, 0x24}, // Tag for expiry date
	}

	// Try to read surname first to test if GET DATA works
	surname, err := getDataByTag(card, tags["surname"])
	if err != nil {
		return nil, err // GET DATA not supported
	}

	// If surname worked, read other fields
	givenName, _ := getDataByTag(card, tags["givenName"])
	personalCode, _ := getDataByTag(card, tags["personalCode"])
	sex, _ := getDataByTag(card, tags["sex"])
	citizenship, _ := getDataByTag(card, tags["citizenship"])
	birthDate, _ := getDataByTag(card, tags["birthDate"])
	documentNr, _ := getDataByTag(card, tags["documentNr"])
	expiryDate, _ := getDataByTag(card, tags["expiryDate"])

	pd := &PersonalData{
		Surname:      sanitizeText(surname),
		GivenName:    sanitizeText(givenName),
		Sex:          sanitizeText(sex),
		Citizenship:  sanitizeText(citizenship),
		BirthDate:    sanitizeText(birthDate),
		PersonalCode: sanitizeText(personalCode),
		DocumentNr:   sanitizeText(documentNr),
		ExpiryDate:   sanitizeText(expiryDate),
		Source:       "chip:GET_DATA",
	}

	return pd, nil
}

// getDataByTag reads data using GET DATA command (ISO 7816-4)
func getDataByTag(card *scard.Card, tag [2]byte) ([]byte, error) {
	// GET DATA: 00 CA P1 P2 Le
	// P1P2 = BER-TLV tag
	cmd := apdu.APDU{CLA: 0x00, INS: 0xCA, P1: tag[0], P2: tag[1], Data: nil, Le: 256}
	resp, err := transmit(card, cmd)
	if err != nil {
		return nil, err
	}
	sw1, sw2 := apdu.StatusWord(resp)

	// Handle SW=6CXX (Wrong Le)
	if sw1 == 0x6C {
		correctLen := int(sw2)
		cmd = apdu.APDU{CLA: 0x00, INS: 0xCA, P1: tag[0], P2: tag[1], Data: nil, Le: correctLen}
		resp, err = transmit(card, cmd)
		if err != nil {
			return nil, err
		}
		sw1, sw2 = apdu.StatusWord(resp)
	}

	if err := apdu.CheckOK(sw1, sw2); err != nil {
		return nil, err
	}

	data := apdu.DataPart(resp)
	// Parse TLV response if present
	if len(data) > 2 && data[0] == tag[0] && data[1] == tag[1] {
		// Skip tag and length bytes
		length := int(data[2])
		if len(data) >= 3+length {
			return data[3 : 3+length], nil
		}
	}

	return data, nil
}
