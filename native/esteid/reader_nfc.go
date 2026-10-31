package esteid

import (
	"errors"
	"fmt"
	"time"

	"ee.aarepuusepp.eid-pd-reader/native/nfc"
	"github.com/ebfe/scard"
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
	return readNFCWithRetry(can, maxAttempts, attemptNFCRead, time.Sleep)
}

// Only connection failures are retried. Once PACE starts, retrying can repeat
// authentication with the same incorrect CAN or hide a secure-messaging error.
type nfcConnectionError struct{ error }

func (e *nfcConnectionError) Unwrap() error { return e.error }

func readNFCWithRetry(can string, maxAttempts int, read func(string, int) (*PersonalData, error), sleep func(time.Duration)) (*PersonalData, error) {
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
			sleep(retryDelays[delayIdx])
		}

		pd, err := read(can, attempt)
		if err == nil {
			return pd, nil
		}

		lastErr = err
		var connectionErr *nfcConnectionError
		if !errors.As(err, &connectionErr) {
			return nil, err
		}
		logVerbose("Attempt %d/%d failed: %v\n", attempt, maxAttempts, err)
	}

	return nil, fmt.Errorf("all NFC read attempts failed: %w", lastErr)
}

// attemptNFCRead performs a single NFC read attempt.
func attemptNFCRead(can string, attemptNum int) (*PersonalData, error) {
	// ConnectAny waits for presence and makes one connection attempt per reader.
	ctx, card, readerName, err := nfc.ConnectAny()
	if err != nil {
		return nil, &nfcConnectionError{fmt.Errorf("NFC connection failed: %w", err)}
	}
	defer ctx.Release()
	defer card.Disconnect(scard.LeaveCard)
	return withNFCCardTransaction(card, func() (*PersonalData, error) {
		return readNFCConnected(card, can, readerName, attemptNum)
	})
}

// PACE and secure messaging share card state across many APDUs. Keep other
// PC/SC clients (including Windows card discovery) out of the entire exchange.
// Unpower before releasing the transaction so no other client inherits our SM
// session. ACR1252 drivers can fail SCARD_RESET_CARD after a contactless read.
func withNFCCardTransaction(card interface {
	BeginTransaction() error
	EndTransaction(scard.Disposition) error
}, read func() (*PersonalData, error)) (pd *PersonalData, err error) {
	if err = card.BeginTransaction(); err != nil {
		return nil, fmt.Errorf("NFC transaction failed: %w", err)
	}
	defer func() {
		if endErr := card.EndTransaction(scard.UnpowerCard); endErr != nil && err == nil {
			pd = nil
			err = fmt.Errorf("NFC transaction cleanup failed: %w", endErr)
		}
	}()
	return read()
}

func readNFCConnected(card *scard.Card, can, readerName string, attemptNum int) (*PersonalData, error) {

	logVerbose("Connected to NFC reader: %s (attempt %d)", readerName, attemptNum)

	// Establish PACE with CAN
	logVerbose("Starting PACE with CAN (length: %d)", len(can))
	sm, err := nfc.StartPACE(card, can)
	if err != nil {
		return nil, fmt.Errorf("PACE failed: %w", err)
	}

	logVerbose("PACE established successfully; Secure Messaging active")
	return readNFCPersonalData(sm)
}

func readNFCPersonalData(sm nfc.SM) (*PersonalData, error) {

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
		return nil, fmt.Errorf("could not select NFC personal data directory")
	}

	// Read personal data fields
	pd := &PersonalData{
		Source: "nfc:PACE+SM",
	}

	// The documented personal-data EFs are selected above. Read them directly;
	// GET DATA probes add six exchanges and can skip document fields on partial results.

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
		logVerbose("  %s: %d bytes", field.name, len(data))
	}

	// If files were accessible but empty, the card has no personal data — not an error
	if filesAccessed > 0 && pd.Surname == "" && pd.GivenName == "" && pd.PersonalCode == "" {
		logVerbose("Card files accessible but contain no personal data (unpersonalized or test card)")
		return pd, nil
	}

	if pd.Surname == "" && pd.GivenName == "" && pd.PersonalCode == "" {
		return nil, fmt.Errorf("failed to read any NFC personal data fields")
	}

	logVerbose("NFC read complete via PACE+SM")

	return pd, nil
}
