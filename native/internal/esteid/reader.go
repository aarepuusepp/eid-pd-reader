package esteid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ebfe/scard"

	"ee.aarepuusepp.eid-pd-reader/native/internal/apdu"
)

// Default file IDs (FID) for EstEID Personal Data (transparent EFs)
var (
	FID_DF_PERSONAL  = [2]byte{0x50, 0x00} // DF '5000'
	FID_SURNAME      = [2]byte{0x50, 0x01} // EF '5001'
	FID_GIVENNAME    = [2]byte{0x50, 0x02} // EF '5002'
	FID_PERSONALCODE = [2]byte{0x50, 0x06} // EF '5006'
)

// PersonalData holds the fields we export
type PersonalData struct {
	Surname      string `json:"surname"`
	GivenName    string `json:"givenName"`
	PersonalCode string `json:"personalCode"`
	Source       string `json:"source"` // "chip:EF500x"
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
	for _, r := range rdrs {
		c, err := ctx.Connect(r, scard.ShareShared, scard.ProtocolAny)
		if err == nil {
			card = c
			break
		}
	}
	if card == nil {
		return nil, fmt.Errorf("pcsc: could not connect to any reader")
	}
	defer card.Disconnect(scard.LeaveCard)

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
	if err := selectDF(card, pdfid); err != nil {
		return nil, fmt.Errorf("select DF 0x%02X%02X: %w", pdfid[0], pdfid[1], err)
	}

	// Read EFs
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

	surname, err := readTransparentEF(card, surnameFID, 256)
	if err != nil {
		return nil, fmt.Errorf("read EF %02X%02X: %w", surnameFID[0], surnameFID[1], err)
	}
	given, err := readTransparentEF(card, givenFID, 256)
	if err != nil {
		return nil, fmt.Errorf("read EF %02X%02X: %w", givenFID[0], givenFID[1], err)
	}
	code, err := readTransparentEF(card, codeFID, 256)
	if err != nil {
		return nil, fmt.Errorf("read EF %02X%02X: %w", codeFID[0], codeFID[1], err)
	}

	pd := &PersonalData{
		Surname:      sanitizeText(surname),
		GivenName:    sanitizeText(given),
		PersonalCode: sanitizeText(code),
		Source:       "chip:EF500x",
	}
	return pd, nil
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
	// 00 A4 01 0C 02 FID (select DF by id)
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x01, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err := transmit(card, cmd)
	if err != nil {
		return err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
}

func selectEF(card *scard.Card, fid [2]byte) error {
	// 00 A4 02 0C 02 FID (select EF by id)
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x02, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	resp, err := transmit(card, cmd)
	if err != nil {
		return err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
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
	if err := apdu.CheckOK(sw1, sw2); err != nil {
		return nil, err
	}
	return bytes.TrimRight(apdu.DataPart(resp), "\x00\x20\n\r"), nil
}

func transmit(card *scard.Card, cmd apdu.APDU) ([]byte, error) {
	resp, err := card.Transmit(cmd.Bytes())
	if err != nil {
		return nil, fmt.Errorf("pcsc: transmit % X: %w", cmd.Bytes(), err)
	}
	return resp, nil
}

func sanitizeText(b []byte) string {
	// EstEID PD tends to be ASCII/UTF-8. Remove trailing control chars.
	out := string(bytes.TrimSpace(b))
	// Collapse internal NULs (some cards may pad with 0x00)
	out = strings.ReplaceAll(out, "\x00", "")
	return out
}

