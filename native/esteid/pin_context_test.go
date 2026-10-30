package esteid

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestPINContextUsesCardIdentityAndChecksRecipient(t *testing.T) {
	card := &x509.Certificate{Raw: []byte{1}, Subject: pkix.Name{
		SerialNumber: "PNOEE-00000000000", CommonName: "fallback",
		Names: []pkix.AttributeTypeAndValue{
			{Type: asn1.ObjectIdentifier{2, 5, 4, 42}, Value: "Mari"},
			{Type: asn1.ObjectIdentifier{2, 5, 4, 4}, Value: "Näidis"},
		},
	}}
	challenge := json.RawMessage(`{"displayContext":{"fileName":"example.cdoc","name":"Spoofed","personalCode":"11111111111"}}`)
	got, err := unwrapPINContext(card, card, challenge, "https://example.test")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Mari Näidis" || got.PersonalCode != "00000000000" || got.FileName != "example.cdoc" || got.Origin != "https://example.test" {
		t.Fatalf("unexpected context: %+v", got)
	}
	if _, err := unwrapPINContext(card, &x509.Certificate{Raw: []byte{2}}, challenge, ""); err == nil {
		t.Fatal("accepted a different card certificate")
	}
}

func TestPINDisplayRemovesSpoofingControlsAndBoundsUnicode(t *testing.T) {
	got := cleanPINDisplay("doc\u202ecdoc\u0000\n", 260)
	if got != "doccdoc" {
		t.Fatalf("unexpected sanitized text %q", got)
	}
	if got := cleanPINDisplay(strings.Repeat("õ", 300), 260); len([]rune(got)) != 260 {
		t.Fatal("wrong Unicode limit")
	}
}

func TestPINTimeoutErrorSurvivesWrapping(t *testing.T) {
	if PINErrorCode(fmt.Errorf("prompt: %w", ErrPINTimeout)) != "ERR_EIDPD_PIN_TIMEOUT" {
		t.Fatal("timeout code lost")
	}
	if PINErrorCode(fmt.Errorf("PIN entry cancelled")) != "" {
		t.Fatal("cancel treated as timeout")
	}
}
