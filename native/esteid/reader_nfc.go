package esteid

import (
	"fmt"
	"time"

	"ee.aarepuusepp.eid-pd-reader/native/nfc"
)

// ReadPersonalDataNFC reads personal data from an Estonian ID card via NFC using PACE + Secure Messaging.
// Requires the 6-digit CAN (Card Access Number) printed on the card.
// Uses default retry count of 2 attempts (initial + 1 retry).
func ReadPersonalDataNFC(can string) (*PersonalData, error) {
	return ReadPersonalDataNFCWithRetry(can, 2)
}

// ReadPersonalDataNFCWithRetry reads personal data from an Estonian ID card via NFC with configurable retry count.
// maxAttempts: total number of attempts (1 = no retry, 2 = 1 retry, etc.)
func ReadPersonalDataNFCWithRetry(can string, maxAttempts int) (*PersonalData, error) {
	// Validate CAN
	if len(can) != 6 {
		return nil, fmt.Errorf("CAN must be exactly 6 digits")
	}
	for _, c := range can {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("CAN must contain only digits")
		}
	}
	
	// Validate maxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	// Try up to maxAttempts times
	// NFC readers sometimes lose connection on first attempt
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			logVerbose("Retrying NFC connection (attempt %d/%d)...", attempt, maxAttempts)
			// Wait 500ms before retry to let the card stabilize
			time.Sleep(500 * time.Millisecond)
		}
		
		pd, err := attemptNFCRead(can, attempt)
		if err == nil {
			return pd, nil
		}
		
		lastErr = err
		logVerbose("Attempt %d failed: %v", attempt, err)
	}
	
	return nil, fmt.Errorf("all NFC read attempts failed: %w", lastErr)
}

// attemptNFCRead performs a single NFC read attempt
func attemptNFCRead(can string, attemptNum int) (*PersonalData, error) {
	// Give NFC reader time to detect card on first attempt
	// NFC readers need a moment to initialize the card
	if attemptNum == 1 {
		time.Sleep(300 * time.Millisecond)
	}
	
	// Connect to NFC reader
	ctx, card, readerName, err := nfc.ConnectAny()
	if err != nil {
		return nil, fmt.Errorf("NFC connection failed: %w", err)
	}
	defer ctx.Release()
	defer card.Disconnect(0)

	logVerbose("Connected to NFC reader: %s (attempt %d)", readerName, attemptNum)

	// Establish PACE with CAN
	logVerbose("Starting PACE with CAN (length: %d)", len(can))
	sm, err := nfc.StartPACE(card, can)
	if err != nil {
		return nil, fmt.Errorf("PACE failed: %w", err)
	}

	logVerbose("PACE established successfully; Secure Messaging active")

	// Correct sequence for Thales cards after PACE (per nfc_sm_debug.md):
	// 1. SELECT AID (PKCS-15 for Thales cards)
	// 2. SELECT DFDD by path
	// 3. SELECT EF 500x + READ BINARY
	
	logVerbose("Selecting PKCS-15 AID via SM...")
	pkcs15AID := []byte{0xA0, 0x00, 0x00, 0x00, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35}
	_, sw, err := sm.Transmit(APDUSelectAID(pkcs15AID))
	if err != nil {
		logVerbose("Warning: SELECT PKCS-15 AID failed: %v (SW=%04X)", err, sw)
	} else if IsSuccess(sw) {
		logVerbose("PKCS-15 AID selected via SM")
	} else {
		logVerbose("SELECT PKCS-15 AID returned SW=%04X, trying EstEID AID...", sw)
		// Try EstEID AID as fallback
		esteidAID := []byte{0xA0, 0x00, 0x00, 0x00, 0x77, 0x01, 0x08, 0x00, 0x07, 0x00, 0x00, 0xFE, 0x00, 0x00, 0x01, 0x00}
		_, sw, err = sm.Transmit(APDUSelectAID(esteidAID))
		if err != nil {
			logVerbose("Warning: SELECT EstEID AID failed: %v (SW=%04X)", err, sw)
		} else if IsSuccess(sw) {
			logVerbose("EstEID AID selected via SM")
		} else {
			logVerbose("SELECT EstEID AID returned SW=%04X, continuing anyway", sw)
		}
	}

	// 2. SELECT DFDD by path (00 A4 08 0C 02 DF DD)
	logVerbose("Selecting DFDD by path via SM...")
	dfddFID := [2]byte{0xDF, 0xDD}
	dfddAPDU := []byte{0x00, 0xA4, 0x08, 0x0C, 0x02, dfddFID[0], dfddFID[1]}
	_, sw, err = sm.Transmit(dfddAPDU)
	if err != nil {
		logVerbose("Warning: SELECT DFDD failed: %v (SW=%04X), continuing anyway", err, sw)
	} else if IsSuccess(sw) {
		logVerbose("DFDD selected via SM")
	} else {
		logVerbose("SELECT DFDD returned SW=%04X, trying DF 5000...", sw)
		// Fallback: try DF 5000
		_, sw, err = sm.Transmit(APDUSelectDF(FID_DF_PERSONAL))
		if err != nil {
			logVerbose("Warning: SELECT DF 5000 failed: %v (SW=%04X)", err, sw)
		} else if IsSuccess(sw) {
			logVerbose("DF 5000 (Personal Data) selected via SM")
		}
	}

	// Read personal data fields
	pd := &PersonalData{
		Source: "nfc:PACE+SM",
	}

	// Try two different approaches for Thales test cards
	
	// Approach 1: Try GET DATA commands (ISO 7816-4)
	logVerbose("Trying GET DATA approach...")
	getDataFields := []struct {
		tag  [2]byte // BER-TLV tag
		name string
		dest *string
	}{
		{[2]byte{0x5F, 0x0E}, "Surname", &pd.Surname},           // Tag 5F0E
		{[2]byte{0x5F, 0x0F}, "Given Name", &pd.GivenName},      // Tag 5F0F
		{[2]byte{0x5F, 0x35}, "Sex", &pd.Sex},                   // Tag 5F35
		{[2]byte{0x5F, 0x28}, "Citizenship", &pd.Citizenship},   // Tag 5F28
		{[2]byte{0x5F, 0x2B}, "Birth Date", &pd.BirthDate},      // Tag 5F2B
		{[2]byte{0x5F, 0x2E}, "Personal Code", &pd.PersonalCode}, // Tag 5F2E (custom)
	}
	
	for _, field := range getDataFields {
		logVerbose("Trying GET DATA for %s (tag=%02X%02X) via SM", field.name, field.tag[0], field.tag[1])
		
		// Build GET DATA APDU: 00 CA <tag> 00
		apdu := []byte{0x00, 0xCA, field.tag[0], field.tag[1], 0x00}
		data, sw, err := sm.Transmit(apdu)
		
		if err == nil && IsSuccess(sw) && len(data) > 0 {
			text := BytesToCleanText(data)
			*field.dest = text
			logVerbose("  %s: %s", field.name, text)
		} else {
			logVerbose("  GET DATA failed for %s: SW=%04X", field.name, sw)
		}
	}
	
	// If GET DATA didn't work, try SELECT+READ approach
	if pd.Surname == "" && pd.GivenName == "" {
		logVerbose("GET DATA approach failed, trying SELECT+READ...")
		
		fields := []struct {
			fid  [2]byte
			name string
			dest *string
		}{
			{FID_SURNAME, "Surname", &pd.Surname},
			{FID_GIVENNAME, "Given Name", &pd.GivenName},
			{FID_SEX, "Sex", &pd.Sex},
			{FID_CITIZENSHIP, "Citizenship", &pd.Citizenship},
			{FID_BIRTHDATE, "Birth Date", &pd.BirthDate},
			{FID_PERSONALCODE, "Personal Code", &pd.PersonalCode},
			{FID_DOCUMENTNR, "Document Number", &pd.DocumentNr},
			{FID_EXPIRYDATE, "Expiry Date", &pd.ExpiryDate},
			{FID_ISSUEDATE, "Issue Date", &pd.IssueDate},
		}

		for _, field := range fields {
			logVerbose("Reading %s (FID=%02X%02X) via SM", field.name, field.fid[0], field.fid[1])

			data, err := nfc.SMSelectAndRead(sm, field.fid, 512)
			if err != nil {
				logVerbose("Warning: failed to read %s: %v", field.name, err)
				continue // Skip failed fields, continue with others
			}

			text := BytesToCleanText(data)
			*field.dest = text
			logVerbose("  %s: %s", field.name, text)
		}
	}

	// Validate that we got at least some critical fields
	if pd.Surname == "" && pd.GivenName == "" && pd.PersonalCode == "" {
		return nil, fmt.Errorf("failed to read any personal data fields (tried GET DATA and SELECT+READ)")
	}

	logVerbose("NFC read complete via PACE+SM")

	return pd, nil
}

