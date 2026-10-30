package esteid

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

func unwrapPINContext(card, recipient *x509.Certificate, challenge json.RawMessage, origin string) (PINContext, error) {
	if card == nil || recipient == nil || !bytes.Equal(card.Raw, recipient.Raw) {
		return PINContext{}, fmt.Errorf("decryption certificate does not match inserted card")
	}
	var metadata struct {
		DisplayContext struct {
			FileName string `json:"fileName"`
		} `json:"displayContext"`
	}
	if err := json.Unmarshal(challenge, &metadata); err != nil {
		return PINContext{}, fmt.Errorf("invalid decryption context")
	}
	var given, surname string
	for _, value := range card.Subject.Names {
		text, ok := value.Value.(string)
		if !ok {
			continue
		}
		switch value.Type.String() {
		case "2.5.4.42":
			given = text
		case "2.5.4.4":
			surname = text
		}
	}
	name := strings.TrimSpace(given + " " + surname)
	if name == "" {
		name = card.Subject.CommonName
	}
	code := strings.TrimPrefix(card.Subject.SerialNumber, "PNOEE-")
	return PINContext{
		FileName:     cleanPINDisplay(metadata.DisplayContext.FileName, 260),
		Name:         cleanPINDisplay(name, 160),
		PersonalCode: cleanPINDisplay(code, 40),
		Origin:       cleanPINDisplay(origin, 200),
	}, nil
}

func cleanPINDisplay(value string, limit int) string {
	var out []rune
	for _, r := range value {
		// Remove controls and bidi formatting that could disguise displayed text.
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			continue
		}
		out = append(out, r)
		if len(out) == limit {
			break
		}
	}
	return strings.TrimSpace(string(out))
}
