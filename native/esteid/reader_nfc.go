package esteid

import (
	"fmt"
	"time"

	"ee.aarepuusepp.eid-pd-reader/native/nfc"
)

// ReadPersonalDataNFC reads personal data from an Estonian ID card via NFC using PACE + Secure Messaging.
// Requires the 6-digit CAN (Card Access Number) printed on the card.
// Uses default retry count of 3 attempts.
func ReadPersonalDataNFC(can string) (*PersonalData, error) {
	return ReadPersonalDataNFCWithRetry(can, 3)
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

	// Try up to maxAttempts times.
	// NFC readers (especially ACR1252) sometimes fail on first power-on
	// when the card isn't fully stable on the RF field.
	retryDelays := []time.Duration{0, 1 * time.Second, 2 * time.Second, 3 * time.Second, 3 * time.Second}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delayIdx := attempt - 1
		if delayIdx >= len(retryDelays) {
			delayIdx = len(retryDelays) - 1
		}
		if attempt > 1 && retryDelays[delayIdx] > 0 {
			logVerbose("Retrying NFC read (attempt %d/%d) in %v...\n", attempt, maxAttempts, retryDelays[delayIdx])
			time.Sleep(retryDelays[delayIdx])
		}

		pd, err := attemptNFCRead(can, attempt)
		if err == nil {
			return pd, nil
		}

		lastErr = err
		logVerbose("Attempt %d/%d failed: %v\n", attempt, maxAttempts, err)
	}
	
	return nil, fmt.Errorf("all NFC read attempts failed: %w", lastErr)
}

// attemptNFCRead performs a single NFC read attempt.
func attemptNFCRead(can string, attemptNum int) (*PersonalData, error) {
	// Connect to NFC reader (ConnectAny waits for card presence and retries power-on)
	ctx, card, readerName, err := nfc.ConnectAny()
	if err != nil {
		return nil, fmt.Errorf("NFC connection failed: %w", err)
	}
	defer ctx.Release()
	defer card.Disconnect(2) // UnpowerCard — drops RF field without triggering a reset IOCTL

	logVerbose("Connected to NFC reader: %s (attempt %d)", readerName, attemptNum)

	// Establish PACE with CAN
	logVerbose("Starting PACE with CAN (length: %d)", len(can))
	sm, err := nfc.StartPACE(card, can)
	if err != nil {
		return nil, fmt.Errorf("PACE failed: %w", err)
	}

	logVerbose("PACE established successfully; Secure Messaging active")

	// Navigate to personal data directory before reading files.
	// IDEMIA cards: personal data files 5001-5009 are under DF 5000 (direct child of MF).
	// Thales cards: personal data files 5001-5009 are under DF.DocumentData (FID DFDD),
	//   which is a subdirectory of the eID application (PKCS-15 AID).
	//   The eID app also has PKCS-15 directory files with overlapping FIDs (5001, 5003, etc.),
	//   so we MUST select DFDD to reach the correct personal data files.

	personalDataDirSelected := false

	// Approach 1: SELECT DF 5000 (works for IDEMIA cards)
	logVerbose("Selecting DF 5000 (Personal Data) via SM...")
	_, sw, err := sm.Transmit(APDUSelectDF(FID_DF_PERSONAL))
	if err == nil && IsSuccess(sw) {
		logVerbose("DF 5000 (Personal Data) selected via SM")
		personalDataDirSelected = true
	} else {
		logVerbose("SELECT DF 5000 failed (SW=%04X), trying Thales eID approach...", sw)

		// Approach 2: Thales path — SELECT PKCS-15 AID, then navigate to DF.DocumentData
		pkcs15AID := []byte{0xA0, 0x00, 0x00, 0x00, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35}
		_, sw, err = sm.Transmit(APDUSelectAID(pkcs15AID))
		if err == nil && IsSuccess(sw) {
			logVerbose("PKCS-15 AID (eID application) selected via SM")

			// Try multiple methods to select DF.DocumentData:
			// Method 1: By AID "Document Data" (446F63756D656E742044617461) per developer guide
			docDataAID := []byte{0x44, 0x6F, 0x63, 0x75, 0x6D, 0x65, 0x6E, 0x74, 0x20, 0x44, 0x61, 0x74, 0x61}
			_, sw, err = sm.Transmit(APDUSelectAID(docDataAID))
			if err == nil && IsSuccess(sw) {
				logVerbose("DF.DocumentData selected by AID via SM")
				personalDataDirSelected = true
			} else {
				logVerbose("SELECT DF.DocumentData by AID failed (SW=%04X), trying by path...", sw)

				// Method 2: By path from current DF (P1=09) — DFDD as child of eID app
				dfddFromDF := []byte{0x00, 0xA4, 0x09, 0x0C, 0x02, 0xDF, 0xDD}
				_, sw, err = sm.Transmit(dfddFromDF)
				if err == nil && IsSuccess(sw) {
					logVerbose("DF.DocumentData (DFDD) selected by path from DF via SM")
					personalDataDirSelected = true
				} else {
					logVerbose("SELECT DFDD from DF failed (SW=%04X), trying absolute path...", sw)

					// Method 3: By absolute path from MF (P1=08)
					dfddFromMF := []byte{0x00, 0xA4, 0x08, 0x0C, 0x02, 0xDF, 0xDD}
					_, sw, err = sm.Transmit(dfddFromMF)
					if err == nil && IsSuccess(sw) {
						logVerbose("DF.DocumentData (DFDD) selected by absolute path via SM")
						personalDataDirSelected = true
					} else {
						logVerbose("SELECT DFDD by absolute path failed (SW=%04X)", sw)
					}
				}
			}
		} else {
			logVerbose("SELECT PKCS-15 AID failed (SW=%04X)", sw)
		}
	}

	if !personalDataDirSelected {
		logVerbose("Warning: could not select personal data directory, will try reading anyway")
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

		filesAccessed := 0
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
				continue
			}

			filesAccessed++
			text := BytesToCleanText(data)
			*field.dest = text
			logVerbose("  %s: %s", field.name, text)
		}

		// If files were accessible but empty, the card has no personal data — not an error
		if filesAccessed > 0 && pd.Surname == "" && pd.GivenName == "" && pd.PersonalCode == "" {
			logVerbose("Card files accessible but contain no personal data (unpersonalized or test card)")
			return pd, nil
		}
	}

	if pd.Surname == "" && pd.GivenName == "" && pd.PersonalCode == "" {
		return nil, fmt.Errorf("failed to read any personal data fields (tried GET DATA and SELECT+READ)")
	}

	logVerbose("NFC read complete via PACE+SM")

	return pd, nil
}

