package esteid

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/aarepuusepp/libcdoc-go/v2/key"
)

// DecryptCertificateResult is returned by get-decrypt-certificate.
type DecryptCertificateResult struct {
	Certificate       string   `json:"certificate"`
	CertificatePEM    string   `json:"certificatePem"`
	ThumbprintSHA256  string   `json:"thumbprintSha256"`
	SupportedKeyTypes []string `json:"supportedKeyTypes"`
}

// PKCS11RuntimeConfig holds resolved PKCS#11 settings for decrypt operations.
type PKCS11RuntimeConfig struct {
	LibraryPath string
	Slot        int
	Label       string
	KeyIDHex    string
	PIN         string
}

// ResolvePKCS11Config merges request overrides with application defaults.
// Public certificate reads pass false and do not require a PIN.
func ResolvePKCS11Config(defaultLibrary string, defaultSlot int, defaultLabel, defaultKeyIDHex string, overrideLibrary string, overrideSlot *int, overrideLabel, overrideKeyIDHex, pin string, promptIfMissingPIN bool) (PKCS11RuntimeConfig, error) {
	cfg := PKCS11RuntimeConfig{
		LibraryPath: strings.TrimSpace(defaultLibrary),
		Slot:        defaultSlot,
		Label:       strings.TrimSpace(defaultLabel),
		KeyIDHex:    strings.TrimSpace(defaultKeyIDHex),
		PIN:         strings.TrimSpace(pin),
	}
	if strings.TrimSpace(overrideLibrary) != "" {
		cfg.LibraryPath = strings.TrimSpace(overrideLibrary)
	}
	if overrideSlot != nil {
		cfg.Slot = *overrideSlot
	}
	if strings.TrimSpace(overrideLabel) != "" {
		cfg.Label = strings.TrimSpace(overrideLabel)
	}
	if strings.TrimSpace(overrideKeyIDHex) != "" {
		cfg.KeyIDHex = strings.TrimSpace(overrideKeyIDHex)
	}
	if cfg.LibraryPath == "" {
		cfg.LibraryPath = DefaultPKCS11LibraryPath()
	}
	if cfg.KeyIDHex == "" {
		cfg.KeyIDHex = DefaultAuthKeyIDHex
	}
	if cfg.LibraryPath == "" {
		return PKCS11RuntimeConfig{}, fmt.Errorf("pkcs11 library path is not configured")
	}
	if cfg.PIN == "" && promptIfMissingPIN {
		entered, err := PromptPIN1("ID-card decryption")
		if err != nil {
			return PKCS11RuntimeConfig{}, err
		}
		cfg.PIN = entered
	}
	if cfg.PIN == "" && promptIfMissingPIN {
		return PKCS11RuntimeConfig{}, fmt.Errorf("pin is required")
	}
	return cfg, nil
}

func (c PKCS11RuntimeConfig) pkcs11Config() (key.PKCS11Config, error) {
	keyID, err := decodeKeyIDHex(c.KeyIDHex)
	if err != nil {
		return key.PKCS11Config{}, err
	}
	return key.PKCS11Config{
		Library: c.LibraryPath,
		Slot:    c.Slot,
		PIN:     c.PIN,
		Label:   c.Label,
		KeyID:   keyID,
	}, nil
}

func decodeKeyIDHex(value string) ([]byte, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "0x"))
	if value == "" {
		return nil, nil
	}
	return hex.DecodeString(value)
}

// GetDecryptCertificate reads the authentication/decryption certificate from the local token.
func GetDecryptCertificate(cfg PKCS11RuntimeConfig) (*DecryptCertificateResult, error) {
	if err := NewCardMonitor().RequireCardPresent(); err != nil {
		return nil, err
	}
	p11, err := cfg.pkcs11Config()
	if err != nil {
		return nil, err
	}
	cert, err := readPublicCertificate(p11)
	if err != nil {
		return nil, err
	}
	return certificateResultFromX509(cert)
}

func certificateResultFromX509(cert *x509.Certificate) (*DecryptCertificateResult, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate not found on token")
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	sum := sha256.Sum256(cert.Raw)
	keyTypes := []string{"RSA"}
	if cert.PublicKeyAlgorithm == x509.ECDSA {
		keyTypes = []string{"EC"}
	}
	return &DecryptCertificateResult{
		Certificate:       base64.StdEncoding.EncodeToString(cert.Raw),
		CertificatePEM:    string(pemBytes),
		ThumbprintSHA256:  hex.EncodeToString(sum[:]),
		SupportedKeyTypes: keyTypes,
	}, nil
}
