package esteid

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"ee.aarepuusepp.eid-pd-reader/native/nfc"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aarepuusepp/libcdoc-go/v2/cdoc"
	"math/big"
	"strings"
	"testing"
	"time"
)

type cryptoSM struct {
	t         *testing.T
	key       *ecdsa.PrivateKey
	thales    bool
	calls     [][]byte
	pinStatus uint16
}

func (s *cryptoSM) Transmit(cmd []byte) ([]byte, uint16, error) {
	s.calls = append(s.calls, bytes.Clone(cmd))
	switch cmd[1] {
	case 0xA4:
		allowed := [][]byte{{0, 0xA4, 0, 0x0C}, {0, 0xA4, 1, 0x0C, 2, 0xAD, 0xF1}}
		if s.thales {
			allowed = [][]byte{{0, 0xA4, 8, 0x0C, 2, 0xAD, 0xF1}}
		}
		valid := false
		for _, candidate := range allowed {
			valid = valid || bytes.Equal(cmd, candidate)
		}
		if !valid {
			s.t.Fatalf("unsupported profile SELECT: %X", cmd)
		}
		return nil, 0x9000, nil
	case 0x20:
		ref, pad := byte(1), byte(0xFF)
		if s.thales {
			ref, pad = 0x81, 0
		}
		want := append([]byte{0, 0x20, 0, ref, 12, '1', '2', '3', '4'}, bytes.Repeat([]byte{pad}, 8)...)
		if !bytes.Equal(cmd, want) {
			s.t.Fatal("incorrect PIN profile encoding")
		}
		if s.pinStatus != 0 {
			return nil, s.pinStatus, nil
		}
		return nil, 0x9000, nil
	case 0x22:
		want := []byte{0, 0x22, 0x41, 0xB8, 9, 0x80, 4, 0xFF, 0x30, 4, 0, 0x84, 1, 0x81}
		if s.thales {
			want = []byte{0, 0x22, 0x41, 0xB8, 3, 0x84, 1, 2}
		}
		if !bytes.Equal(cmd, want) {
			s.t.Fatal("incorrect decipher environment")
		}
		return nil, 0x9000, nil
	case 0x2A:
		if len(cmd) != 104 || cmd[5] != 0 || cmd[3] != 0x86 {
			s.t.Fatal("incorrect decipher APDU")
		}
		peer, err := ecdsa.ParseUncompressedPublicKey(elliptic.P384(), cmd[6:103])
		if err != nil {
			return nil, 0, err
		}
		private, err := s.key.ECDH()
		if err != nil {
			return nil, 0, err
		}
		public, err := peer.ECDH()
		if err != nil {
			return nil, 0, err
		}
		shared, err := private.ECDH(public)
		return shared, 0x9000, err
	}
	return nil, 0, fmt.Errorf("unexpected command")
}
func nfcTestCert(t *testing.T) (*ecdsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(123), Subject: pkix.Name{CommonName: "NFC Test"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageKeyAgreement | x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return key, cert
}

// End-to-end CDOC builder -> challenge -> APDU adapter -> CDOC decrypt.
// The simulated card calculates ECDH with crypto/ecdh, independently of our adapter.
func TestNFCTokenCDOCRoundTrip(t *testing.T) {
	key, cert := nfcTestCert(t)
	for _, version := range []cdoc.Version{cdoc.Version11, cdoc.Version20} {
		for _, thales := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/thales=%v", version, thales), func(t *testing.T) {
				b := cdoc.NewBuilder().Version(version)
				plaintext := []byte("NFC round trip")
				b.WithDataFile("test.txt", bytes.NewReader(plaintext), int64(len(plaintext)))
				if err := b.WithRecipientCertificate(cert); err != nil {
					t.Fatal(err)
				}
				var buf bytes.Buffer
				if err := b.BuildToWriter(&buf); err != nil {
					t.Fatal(err)
				}
				challenge, err := cdoc.PrepareClientUnwrap(buf.Bytes(), cert)
				if err != nil {
					t.Fatal(err)
				}
				sm := &cryptoSM{t: t, key: key, thales: thales}
				ref := byte(0x81)
				if thales {
					ref = 2
				}
				token := &nfcCryptoCard{sm: sm, cert: cert, thales: thales, keyRef: ref, pin: []byte("1234")}
				cek, err := cdoc.UnwrapTransportKeyWithToken(token, *challenge)
				if err != nil {
					t.Fatal(err)
				}
				defer clear(cek)
				files, err := cdoc.FinishClientDecrypt(buf.Bytes(), cert, cek)
				if err != nil || len(files) != 1 || !bytes.Equal(files[0].Content, plaintext) {
					t.Fatalf("round trip failed: %v", err)
				}
				if token.pin != nil {
					t.Fatal("PIN retained after verify")
				}
				wantCalls := 5
				if thales {
					wantCalls = 4
				}
				if len(sm.calls) != wantCalls {
					t.Fatal("unexpected repeated card operation")
				}
			})
		}
	}
}
func TestNFCPINFailureNeverRetriesOrDeciphers(t *testing.T) {
	key, cert := nfcTestCert(t)
	sm := &cryptoSM{t: t, key: key, pinStatus: 0x63C2}
	token := &nfcCryptoCard{sm: sm, cert: cert, keyRef: 0x81, pin: []byte("1234")}
	if _, err := token.DecryptECSharedSecret(&key.PublicKey); err == nil {
		t.Fatal("incorrect PIN accepted")
	}
	if len(sm.calls) != 3 || token.pin != nil {
		t.Fatal("PIN was retained or retried")
	}
}

type nfcSelectStep struct {
	command, response []byte
	status            uint16
}
type nfcSelectScript struct {
	t     *testing.T
	steps []nfcSelectStep
}

func TestNFCSelectionFailuresStopBeforePIN(t *testing.T) {
	pkcs15 := []byte{0, 0xA4, 4, 0, 12, 0xA0, 0, 0, 0, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35, 0}
	for _, stage := range []string{"security", "master file", "authentication DF ADF1", "authentication certificate EF 3401"} {
		t.Run(stage, func(t *testing.T) {
			s := &nfcSelectScript{t: t, steps: []nfcSelectStep{{pkcs15, nil, 0x6A82}}}
			if stage == "security" {
				s.steps[0].status = 0x6982
			} else {
				steps := []nfcSelectStep{
					{[]byte{0, 0xA4, 0, 0x0C}, nil, 0x9000},
					{[]byte{0, 0xA4, 1, 0x0C, 2, 0xAD, 0xF1}, nil, 0x9000},
					{[]byte{0, 0xA4, 2, 0x0C, 2, 0x34, 1}, nil, 0x9000},
				}
				n := map[string]int{"master file": 1, "authentication DF ADF1": 2, "authentication certificate EF 3401": 3}[stage]
				steps[n-1].status = 0x6A82
				s.steps = append(s.steps, steps[:n]...)
			}
			_, err := selectNFCCryptoCard(s)
			if err == nil || len(s.steps) != 0 {
				t.Fatal("selection error was ignored")
			}
			if stage != "security" && !strings.Contains(err.Error(), stage) {
				t.Fatalf("missing failure stage: %v", err)
			}
		})
	}
}

func (s *nfcSelectScript) Transmit(cmd []byte) ([]byte, uint16, error) {
	s.t.Helper()
	if len(s.steps) == 0 {
		s.t.Fatalf("unexpected command %X", cmd)
	}
	next := s.steps[0]
	s.steps = s.steps[1:]
	if !bytes.Equal(cmd, next.command) {
		s.t.Fatalf("SELECT/read mismatch: got %X, want %X", cmd, next.command)
	}
	return next.response, next.status, nil
}

// Full certificate selection, including strict SELECT parameters from the card
// guides. The earlier token-only tests skipped this path and missed SW=6A86.
func TestNFCCertificateSelectionProfiles(t *testing.T) {
	_, cert := nfcTestCert(t)
	for _, thales := range []bool{false, true} {
		t.Run(fmt.Sprintf("thales=%v", thales), func(t *testing.T) {
			s := &nfcSelectScript{t: t}
			add := func(cmd, response []byte, sw uint16) { s.steps = append(s.steps, nfcSelectStep{cmd, response, sw}) }
			read := func(data []byte) {
				for offset := 0; offset < len(data); offset += nfc.MaxSMLe {
					end := min(offset+nfc.MaxSMLe, len(data))
					sw := uint16(0x9000)
					if end == len(data) {
						sw = 0x6282
					}
					add([]byte{0, 0xB0, byte(offset >> 8), byte(offset), nfc.MaxSMLe}, data[offset:end], sw)
				}
			}
			pkcs15 := []byte{0, 0xA4, 4, 0, 12, 0xA0, 0, 0, 0, 0x63, 0x50, 0x4B, 0x43, 0x53, 0x2D, 0x31, 0x35, 0}
			if thales {
				add(pkcs15, []byte{0x6F, 0}, 0x9000)
				add([]byte{0, 0xA4, 8, 0x0C, 2, 0x50, 1}, nil, 0x9000)
				// Minimal DER PKCS#15 EC private-key object: authId=01,
				// derive usage bit, keyReference=2 (no production parser helper).
				read([]byte{0xA0, 0x1F, 0x30, 0x0A, 0x0C, 1, 'a', 3, 2, 6, 0xC0, 4, 1, 1,
					0x30, 0x0F, 4, 1, 0x12, 3, 3, 7, 0x20, 0x80, 3, 2, 3, 0xB8, 2, 1, 2, 0x30, 0})
				add([]byte{0, 0xA4, 8, 0x0C, 4, 0xAD, 0xF1, 0x34, 0x11}, nil, 0x9000)
			} else {
				add(pkcs15, nil, 0x6A86)
				// ID1 guide 2018 p25 / 2025 p95: no AID prerequisite.
				// A card need not expose the newer MF AID to support this flow.
				add([]byte{0, 0xA4, 0, 0x0C}, nil, 0x9000)
				add([]byte{0, 0xA4, 1, 0x0C, 2, 0xAD, 0xF1}, nil, 0x9000)
				add([]byte{0, 0xA4, 2, 0x0C, 2, 0x34, 1}, nil, 0x9000)
			}
			read(cert.Raw)
			card, err := selectNFCCryptoCard(s)
			if err != nil {
				t.Fatal(err)
			}
			if len(s.steps) != 0 || card.thales != thales || !bytes.Equal(card.cert.Raw, cert.Raw) {
				t.Fatal("incomplete or incorrect certificate selection")
			}
		})
	}
}
func TestNFCRejectsMalformedPointBeforePIN(t *testing.T) {
	sm := &cryptoSM{t: t}
	token := &nfcCryptoCard{sm: sm, pin: []byte("1234")}
	for _, p := range []*ecdsa.PublicKey{nil, {Curve: elliptic.P384(), X: big.NewInt(1), Y: big.NewInt(1)}} {
		if _, err := token.DecryptECSharedSecret(p); err == nil {
			t.Fatal("invalid point accepted")
		}
	}
	if len(sm.calls) != 0 {
		t.Fatal("invalid request touched card")
	}
}
func TestThalesActiveKeyDirectory(t *testing.T) {
	marshal := func(v any) []byte {
		b, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	seq := func(parts ...[]byte) []byte {
		return marshal(asn1.RawValue{Tag: 16, IsCompound: true, Bytes: bytes.Join(parts, nil)})
	}
	entry := func(ref int) []byte {
		common := seq(marshal("authentication key"), marshal(asn1.BitString{Bytes: []byte{0xC0}, BitLength: 2}), marshal([]byte{1}))
		attrs := seq(marshal([]byte{0x12}), marshal(asn1.BitString{Bytes: []byte{0x20, 0x80}, BitLength: 9}), marshal(asn1.BitString{Bytes: []byte{0xB8}, BitLength: 5}), marshal(ref))
		return marshal(asn1.RawValue{Class: 2, Tag: 0, IsCompound: true, Bytes: bytes.Join([][]byte{common, attrs, seq()}, nil)})
	}
	for _, ref := range []int{1, 2} {
		got, err := thalesAuthenticationKey(entry(ref))
		if err != nil || got != byte(ref) {
			t.Fatalf("%d %v", got, err)
		}
	}
	for _, input := range [][]byte{nil, {0xA0, 0xFF}, entry(3), append(entry(1), entry(2)...)} {
		if _, err := thalesAuthenticationKey(input); err == nil {
			t.Fatal("invalid/ambiguous key accepted")
		}
	}
}

func TestNFCUnwrapCardSwapAndPINTimeout(t *testing.T) {
	_, cert := nfcTestCert(t)
	_, other := nfcTestCert(t)
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
	raw, _ := json.Marshal(map[string]any{"recipientCertPem": certPEM, "displayContext": map[string]string{"fileName": "example.cdoc"}})
	for _, scenario := range []string{"swap", "wrong-card", "timeout"} {
		t.Run(scenario, func(t *testing.T) {
			sessions, prompts := 0, 0
			sm := &cryptoSM{t: t}
			session := func(_ string, op func(*nfcCryptoCard) error) error {
				sessions++
				actual := cert
				if scenario == "wrong-card" || sessions == 2 {
					actual = other
				}
				return op(&nfcCryptoCard{cert: actual, sm: sm})
			}
			prompt := func(_ string, context PINContext) (string, error) {
				prompts++
				if context.Name != "NFC Test" || context.FileName != "example.cdoc" {
					t.Fatal("untrusted PIN context")
				}
				if scenario == "timeout" {
					return "", ErrPINTimeout
				}
				return "1234", nil
			}
			result, err := unwrapTransportKeyNFCWith("123456", raw, certPEM, "https://digisign.test", session, prompt)
			if err == nil || result != nil || len(sm.calls) != 0 {
				t.Fatal("unsafe card operation")
			}
			if scenario == "wrong-card" && prompts != 0 {
				t.Fatal("PIN prompted for wrong card")
			}
			if scenario == "timeout" && (!errors.Is(err, ErrPINTimeout) || sessions != 1) {
				t.Fatal("timeout did not stop card operation")
			}
			if scenario == "swap" && (sessions != 2 || prompts != 1) {
				t.Fatal("card not rechecked after PIN")
			}
		})
	}
}
