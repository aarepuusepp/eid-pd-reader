package esteid

// Card profiles: ID1 Developer Guide 2025 sections 9.7/10 (IDEMIA),
// GED24035_015 release 1.33 sections 5, 7.6, 7.10, 8.3 (Thales).
// Only the documented P-384 authentication keys are supported. No PIN guessing,
// automatic PIN retries, plaintext APDUs, or exported private keys are used.
import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"ee.aarepuusepp.eid-pd-reader/native/nfc"
	"github.com/aarepuusepp/libcdoc-go/v2/cdoc"
	cdoccrypto "github.com/aarepuusepp/libcdoc-go/v2/crypto"
	"github.com/ebfe/scard"
)

type nfcCryptoCard struct {
	sm     nfc.SM
	cert   *x509.Certificate
	thales bool
	keyRef byte
	pin    []byte
}

func validNFCCAN(can string) bool {
	if len(can) != 6 {
		return false
	}
	for _, c := range can {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Retry only power-on failures, never PACE, PIN verification or private operations.
// The PIN window is opened between sessions, not while holding a PC/SC transaction.
func withNFCCrypto(can string, operation func(*nfcCryptoCard) error) (err error) {
	if !validNFCCAN(can) {
		return fmt.Errorf("INVALID_REQUEST: CAN must contain six digits")
	}
	var ctx *scard.Context
	var card *scard.Card
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		ctx, card, _, err = nfc.ConnectAny()
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("CARD_NOT_FOUND: NFC connection failed: %w", err)
	}
	defer ctx.Release()
	defer card.Disconnect(scard.LeaveCard)
	if err = card.BeginTransaction(); err != nil {
		return fmt.Errorf("NFC transaction failed: %w", err)
	}
	defer func() {
		if e := card.EndTransaction(scard.UnpowerCard); e != nil && err == nil {
			err = fmt.Errorf("NFC cleanup failed: %w", e)
		}
	}()
	sm, err := nfc.StartPACE(card, can)
	if err != nil {
		return fmt.Errorf("NFC PACE failed: %w", err)
	}
	selected, err := selectNFCCryptoCard(sm)
	if err != nil {
		return err
	}
	return operation(selected)
}

func smCommand(sm nfc.SM, cmd []byte) ([]byte, error) {
	data, sw, err := sm.Transmit(cmd)
	if err != nil {
		return nil, err
	}
	if sw != 0x9000 {
		return nil, fmt.Errorf("NFC command %02X failed: SW=%04X", cmd[1], sw)
	}
	return data, nil
}

func selectNFCCryptoCard(sm nfc.SM) (*nfcCryptoCard, error) {
	// Detect a documented application before choosing a PIN reference/padding.
	pkcs15 := []byte{0xA0, 0, 0, 0, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35}
	// Thales SELECT requests FCI and must include Le (guide section 7.1).
	_, sw, err := sm.Transmit(append(APDUSelectAID(pkcs15), 0))
	if err != nil {
		return nil, err
	}
	c := &nfcCryptoCard{sm: sm, thales: sw == 0x9000, keyRef: 0x81}
	if c.thales {
		// Active authentication key references can change after card re-keying.
		directory, err := readNFCPath(sm, []byte{0x50, 0x01})
		if err != nil {
			return nil, err
		}
		c.keyRef, err = thalesAuthenticationKey(directory)
		if err != nil {
			return nil, err
		}
	} else {
		if sw != 0x6A82 && sw != 0x6A86 && sw != 0x6A88 {
			return nil, fmt.Errorf("NFC application selection failed: SW=%04X", sw)
		}
		// IAS-ECC is already available after PACE. Follow the certificate-read
		// sequence from both ID1 guides: MF -> ADF1 -> EF 3401. Requiring the
		// newer guide's MF AID first rejects cards that return 6A82 for that AID.
	}
	fid := byte(0x01)
	if c.thales {
		fid = 0x11
	}
	var der []byte
	if c.thales {
		der, err = readNFCPath(sm, []byte{0xAD, 0xF1, 0x34, fid})
	} else {
		// IAS-ECC supports child DF/EF selection, not absolute-path P1=08.
		err = c.selectAuthenticationDF()
		if err == nil {
			_, err = smCommand(sm, APDUSelectEF([2]byte{0x34, fid}))
			if err != nil {
				err = fmt.Errorf("NFC select authentication certificate EF 3401: %w", err)
			}
		}
		if err == nil {
			der, err = nfc.SMReadEF(sm, 8192)
		}
	}
	if err != nil {
		return nil, err
	}
	// Files may have padding after the DER object. Only parse that first object.
	var raw asn1.RawValue
	if _, err = asn1.Unmarshal(der, &raw); err != nil {
		return nil, fmt.Errorf("invalid NFC certificate: %w", err)
	}
	c.cert, err = x509.ParseCertificate(raw.FullBytes)
	if err != nil {
		return nil, err
	}
	pub, ok := c.cert.PublicKey.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P384() || c.cert.KeyUsage&x509.KeyUsageKeyAgreement == 0 {
		return nil, fmt.Errorf("unsupported NFC authentication certificate: P-384 key agreement required")
	}
	return c, nil
}

func readNFCPath(sm nfc.SM, path []byte) ([]byte, error) {
	cmd := append([]byte{0, 0xA4, 0x08, 0x0C, byte(len(path))}, path...)
	if _, err := smCommand(sm, cmd); err != nil {
		return nil, err
	}
	return nfc.SMReadEF(sm, 8192)
}

func (c *nfcCryptoCard) selectAuthenticationDF() error {
	if c.thales {
		_, err := smCommand(c.sm, []byte{0, 0xA4, 0x08, 0x0C, 2, 0xAD, 0xF1})
		return err
	}
	// Return to MF explicitly: certificate reading left us inside ADF1.
	// ID1 guide (2018 and 2025) uses a case-1 SELECT MF, without AID/FID.
	if _, err := smCommand(c.sm, []byte{0, 0xA4, 0, 0x0C}); err != nil {
		return fmt.Errorf("NFC select master file: %w", err)
	}
	_, err := smCommand(c.sm, APDUSelectDF([2]byte{0xAD, 0xF1}))
	if err != nil {
		return fmt.Errorf("NFC select authentication DF ADF1: %w", err)
	}
	return err
}

// Read the PKCS#15 private-key directory, not a guessed active key slot.
func thalesAuthenticationKey(der []byte) (byte, error) {
	var found byte
	for len(der) > 0 {
		var entry asn1.RawValue
		rest, err := asn1.Unmarshal(der, &entry)
		if err != nil {
			return 0, fmt.Errorf("invalid NFC key directory")
		}
		der = rest
		if entry.Class != 2 || entry.Tag != 0 {
			continue
		}
		fields, err := rawFields(entry.Bytes)
		if err != nil || len(fields) < 3 {
			return 0, fmt.Errorf("invalid NFC key object")
		}
		common, err := rawFields(fields[0].Bytes)
		if err != nil {
			return 0, err
		}
		auth := false
		for _, v := range common {
			if v.Class == 0 && v.Tag == 4 && bytes.Equal(v.Bytes, []byte{1}) {
				auth = true
			}
		}
		if !auth {
			continue
		}
		attrs, err := rawFields(fields[1].Bytes)
		if err != nil || len(attrs) < 4 {
			return 0, fmt.Errorf("invalid NFC authentication key")
		}
		var usage asn1.BitString
		var ref int
		if _, err = asn1.Unmarshal(attrs[1].FullBytes, &usage); err != nil || usage.At(8) != 1 {
			continue
		}
		if _, err = asn1.Unmarshal(attrs[3].FullBytes, &ref); err != nil || ref < 1 || ref > 2 || found != 0 {
			return 0, fmt.Errorf("unsupported or ambiguous NFC authentication key")
		}
		found = byte(ref)
	}
	if found == 0 {
		return 0, fmt.Errorf("NFC authentication key not found")
	}
	return found, nil
}
func rawFields(der []byte) ([]asn1.RawValue, error) {
	var fields []asn1.RawValue
	for len(der) > 0 {
		var v asn1.RawValue
		rest, err := asn1.Unmarshal(der, &v)
		if err != nil {
			return nil, err
		}
		fields = append(fields, v)
		der = rest
	}
	return fields, nil
}

func GetDecryptCertificateNFC(can string) (*DecryptCertificateResult, error) {
	var result *DecryptCertificateResult
	err := withNFCCrypto(can, func(c *nfcCryptoCard) error {
		var err error
		result, err = certificateResultFromX509(c.cert)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func UnwrapTransportKeyNFC(can string, challengeJSON json.RawMessage, certificate, origin string) (*UnwrapTransportKeyResult, error) {
	return unwrapTransportKeyNFCWith(can, challengeJSON, certificate, origin, withNFCCrypto, promptPIN1WithContext)
}

func unwrapTransportKeyNFCWith(can string, challengeJSON json.RawMessage, certificate, origin string,
	session func(string, func(*nfcCryptoCard) error) error,
	prompt func(string, PINContext) (string, error)) (*UnwrapTransportKeyResult, error) {
	var challenge cdoc.ClientUnwrapChallenge
	if err := json.Unmarshal(challengeJSON, &challenge); err != nil {
		return nil, fmt.Errorf("invalid challenge: %w", err)
	}
	if challenge.RecipientCertPEM == "" {
		challenge.RecipientCertPEM = certificate
	}
	recipient, err := challenge.RecipientCertificate()
	if err != nil {
		return nil, err
	}
	var details PINContext
	err = session(can, func(c *nfcCryptoCard) error {
		var e error
		details, e = unwrapPINContext(c.cert, recipient, challengeJSON, origin)
		return e
	})
	if err != nil {
		return nil, err
	}
	pin, err := prompt("ID-card NFC decryption", details)
	if err != nil {
		return nil, err
	}
	defer func() { pin = "" }()
	var cek []byte
	defer func() { clear(cek) }()
	err = session(can, func(c *nfcCryptoCard) error {
		// Recheck after PIN entry: the card may have been swapped while the dialog was open.
		if !bytes.Equal(c.cert.Raw, recipient.Raw) {
			return fmt.Errorf("decryption certificate does not match inserted card")
		}
		// The Token validates the challenge/point before its first and only PIN attempt.
		c.pin = []byte(pin)
		defer clear(c.pin)
		var e error
		cek, e = cdoc.UnwrapTransportKeyWithToken(c, challenge)
		return e
	})
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(cek)
	return &UnwrapTransportKeyResult{ContentEncryptionKeyBase64: b64, ContentEncryptionKey: b64}, nil
}

func (c *nfcCryptoCard) GetCertificate() (*x509.Certificate, error) { return c.cert, nil }
func (c *nfcCryptoCard) DecryptRSA([]byte) ([]byte, error) {
	return nil, fmt.Errorf("RSA decryption is not supported over NFC")
}
func (c *nfcCryptoCard) DecryptEC(p *ecdsa.PublicKey, a, u, v, k []byte) ([]byte, error) {
	return c.DecryptECWithDigest(p, a, u, v, k, "")
}
func (c *nfcCryptoCard) DecryptECWithDigest(p *ecdsa.PublicKey, a, u, v, k []byte, digest string) ([]byte, error) {
	if digest == "" {
		digest = cdoccrypto.DigestSHA384URI
	}
	// Validate KDF inputs before asking the card to consume a PIN attempt.
	check, err := cdoccrypto.ConcatKDFWithDigest(make([]byte, 48), a, u, v, digest, 32)
	clear(check)
	if err != nil {
		return nil, err
	}
	if len(k) < 24 || len(k)%8 != 0 {
		return nil, fmt.Errorf("invalid wrapped key")
	}
	shared, err := c.DecryptECSharedSecret(p)
	if err != nil {
		return nil, err
	}
	defer clear(shared)
	wrap, err := cdoccrypto.ConcatKDFWithDigest(shared, a, u, v, digest, 32)
	if err != nil {
		return nil, err
	}
	defer clear(wrap)
	return cdoccrypto.UnwrapAES256(k, wrap)
}

func (c *nfcCryptoCard) DecryptECSharedSecret(p *ecdsa.PublicKey) ([]byte, error) {
	if p == nil || p.Curve != elliptic.P384() || p.X == nil || p.Y == nil || !p.Curve.IsOnCurve(p.X, p.Y) {
		return nil, fmt.Errorf("invalid NFC ECDH point")
	}
	if len(c.pin) < 4 || len(c.pin) > 12 {
		return nil, fmt.Errorf("invalid PIN1 length")
	}
	for _, v := range c.pin {
		if v < '0' || v > '9' {
			return nil, fmt.Errorf("PIN1 must contain digits")
		}
	}
	// Select authentication DF before VERIFY, never reselect it afterwards.
	if err := c.selectAuthenticationDF(); err != nil {
		return nil, err
	}
	ref, pad := byte(1), byte(0xFF)
	if c.thales {
		ref, pad = 0x81, 0
	}
	verify := append([]byte{0, 0x20, 0, ref, 12}, bytes.Repeat([]byte{pad}, 12)...)
	copy(verify[5:], c.pin)
	_, sw, err := c.sm.Transmit(verify)
	clear(verify)
	clear(c.pin)
	c.pin = nil
	if err != nil {
		return nil, err
	}
	if sw&0xFFF0 == 0x63C0 {
		return nil, fmt.Errorf("PIN1 verification failed; remaining attempts: %d", sw&15)
	}
	if sw != 0x9000 {
		return nil, fmt.Errorf("PIN1 verification failed: SW=%04X", sw)
	}
	mse := []byte{0, 0x22, 0x41, 0xB8, 9, 0x80, 4, 0xFF, 0x30, 4, 0, 0x84, 1, c.keyRef}
	if c.thales {
		mse = []byte{0, 0x22, 0x41, 0xB8, 3, 0x84, 1, c.keyRef}
	}
	if _, err = smCommand(c.sm, mse); err != nil {
		return nil, err
	}
	point := elliptic.Marshal(p.Curve, p.X, p.Y)
	cmd := append([]byte{0, 0x2A, 0x80, 0x86, byte(len(point) + 1), 0}, point...)
	cmd = append(cmd, 0)
	shared, err := smCommand(c.sm, cmd)
	if err != nil {
		return nil, err
	}
	if len(shared) != 48 {
		clear(shared)
		return nil, fmt.Errorf("invalid NFC ECDH secret length")
	}
	return shared, nil
}
