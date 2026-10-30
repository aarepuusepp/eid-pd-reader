package esteid

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/aarepuusepp/libcdoc-go/v2/cdoc"
	"github.com/aarepuusepp/libcdoc-go/v2/key"
)

// DecryptOptions describes a single local PKCS#11-backed container decryption.
type DecryptOptions struct {
	ContainerBase64 string
	PKCS11Library   string
	PIN             string
	Slot            int
	Label           string
	KeyIDHex        string
}

// DecryptedFile is returned to the browser extension. Content is base64 encoded
// so native messaging can keep using JSON safely for binary files.
type DecryptedFile struct {
	Name          string `json:"name"`
	ContentBase64 string `json:"contentBase64"`
	Size          int    `json:"size"`
}

// DecryptResult contains files decrypted from a CDOC container.
type DecryptResult struct {
	Files       []DecryptedFile `json:"files"`
	FileCount   int             `json:"fileCount"`
	TokenSource string          `json:"tokenSource"`
}

// DecryptContainer decrypts a CDOC container with the authentication/decryption
// key available through a local PKCS#11 module.
func DecryptContainer(opts DecryptOptions) (*DecryptResult, error) {
	if strings.TrimSpace(opts.ContainerBase64) == "" {
		return nil, fmt.Errorf("containerBase64 is required")
	}
	if strings.TrimSpace(opts.PKCS11Library) == "" {
		return nil, fmt.Errorf("pkcs11.libraryPath is required")
	}
	if strings.TrimSpace(opts.PIN) == "" {
		return nil, fmt.Errorf("pin is required")
	}
	if opts.Slot < 0 {
		opts.Slot = 0
	}

	container, err := base64.StdEncoding.DecodeString(opts.ContainerBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid containerBase64: %w", err)
	}

	var keyID []byte
	if strings.TrimSpace(opts.KeyIDHex) != "" {
		keyID, err = hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(opts.KeyIDHex), "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid pkcs11.keyIdHex: %w", err)
		}
	}

	token, err := cdoc.NewPKCS11Token(key.PKCS11Config{
		Library: strings.TrimSpace(opts.PKCS11Library),
		Slot:    opts.Slot,
		PIN:     opts.PIN,
		Label:   strings.TrimSpace(opts.Label),
		KeyID:   keyID,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = token.Close() }()

	cert, err := token.GetCertificate()
	if err != nil {
		return nil, err
	}

	files, err := cdoc.NewDecrypter().
		WithToken(token).
		WithCertificate(cert).
		WithCDOC(bytes.NewReader(container)).
		Decrypt()
	if err != nil {
		return nil, err
	}

	result := &DecryptResult{
		Files:       make([]DecryptedFile, 0, len(files)),
		FileCount:   len(files),
		TokenSource: "PKCS11",
	}
	for _, f := range files {
		result.Files = append(result.Files, DecryptedFile{
			Name:          f.Name,
			ContentBase64: base64.StdEncoding.EncodeToString(f.Content),
			Size:          len(f.Content),
		})
	}
	return result, nil
}
