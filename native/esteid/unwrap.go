package esteid

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aarepuusepp/libcdoc-go/v2/cdoc"
)

// UnwrapTransportKeyResult contains the content encryption key as base64.
type UnwrapTransportKeyResult struct {
	ContentEncryptionKeyBase64 string `json:"contentEncryptionKeyBase64"`
	ContentEncryptionKey       string `json:"contentEncryptionKey"`
}

// UnwrapTransportKey unwraps a CDOC 1.x transport key using the local PKCS#11 token.
func UnwrapTransportKey(cfg PKCS11RuntimeConfig, challengeJSON json.RawMessage, certificate string, origin string) (*UnwrapTransportKeyResult, error) {
	var challenge cdoc.ClientUnwrapChallenge
	if len(challengeJSON) > 0 {
		if err := json.Unmarshal(challengeJSON, &challenge); err != nil {
			return nil, fmt.Errorf("invalid challenge: %w", err)
		}
	}
	if challenge.RecipientCertPEM == "" && certificate != "" {
		challenge.RecipientCertPEM = certificate
	}
	p11, err := cfg.pkcs11Config()
	if err != nil {
		return nil, err
	}
	// Identity comes from the actual selected token, never page-supplied text.
	cardCert, err := readPublicCertificate(p11)
	if err != nil {
		return nil, err
	}
	recipientCert, err := challenge.RecipientCertificate()
	if err != nil {
		return nil, err
	}
	details, err := unwrapPINContext(cardCert, recipientCert, challengeJSON, origin)
	if err != nil {
		return nil, err
	}
	if cfg.PIN == "" {
		p11.PIN, err = promptPIN1WithContext("ID-card decryption", details)
		if err != nil {
			return nil, err
		}
	}
	cek, err := cdoc.UnwrapTransportKeyWithPKCS11(p11, challenge)
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(cek)
	return &UnwrapTransportKeyResult{
		ContentEncryptionKeyBase64: b64,
		ContentEncryptionKey:       b64,
	}, nil
}
