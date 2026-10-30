//go:build cgo

package esteid

import (
	"crypto/x509"
	"fmt"

	"github.com/aarepuusepp/libcdoc-go/v2/key"
	"github.com/miekg/pkcs11"
)

// This deliberately excludes Login and private-key operations: the certificate
// is public, and the separate unwrap command owns the PIN-authenticated session.
type publicCertificateModule interface {
	Initialize(...pkcs11.InitializeOption) error
	Finalize() error
	Destroy()
	GetSlotList(bool) ([]uint, error)
	OpenSession(uint, uint) (pkcs11.SessionHandle, error)
	CloseSession(pkcs11.SessionHandle) error
	FindObjectsInit(pkcs11.SessionHandle, []*pkcs11.Attribute) error
	FindObjects(pkcs11.SessionHandle, int) ([]pkcs11.ObjectHandle, bool, error)
	FindObjectsFinal(pkcs11.SessionHandle) error
	GetAttributeValue(pkcs11.SessionHandle, pkcs11.ObjectHandle, []*pkcs11.Attribute) ([]*pkcs11.Attribute, error)
}

func readPublicCertificate(cfg key.PKCS11Config) (*x509.Certificate, error) {
	module := pkcs11.New(cfg.Library)
	if module == nil {
		return nil, fmt.Errorf("failed to load PKCS#11 library")
	}
	return readPublicCertificateFromModule(module, cfg)
}

func readPublicCertificateFromModule(module publicCertificateModule, cfg key.PKCS11Config) (*x509.Certificate, error) {
	defer module.Destroy()
	if err := module.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize PKCS#11 library: %w", err)
	}
	defer func() { _ = module.Finalize() }()
	slots, err := module.GetSlotList(true)
	if err != nil {
		return nil, fmt.Errorf("list PKCS#11 slots: %w", err)
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("no slots with tokens found")
	}
	if cfg.Slot < 0 {
		return nil, fmt.Errorf("invalid PKCS#11 slot index")
	}
	// Match libcdoc-go's slot-index selection, including its first-slot fallback.
	index := cfg.Slot
	if index >= len(slots) {
		index = 0
	}
	session, err := module.OpenSession(slots[index], pkcs11.CKF_SERIAL_SESSION)
	if err != nil {
		return nil, fmt.Errorf("open public PKCS#11 session: %w", err)
	}
	defer func() { _ = module.CloseSession(session) }()
	attributes := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_CERTIFICATE),
		pkcs11.NewAttribute(pkcs11.CKA_CERTIFICATE_TYPE, pkcs11.CKC_X_509),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
	}
	if len(cfg.KeyID) > 0 {
		attributes = append(attributes, pkcs11.NewAttribute(pkcs11.CKA_ID, cfg.KeyID))
	}
	if cfg.Label != "" {
		attributes = append(attributes, pkcs11.NewAttribute(pkcs11.CKA_LABEL, cfg.Label))
	}
	if err := module.FindObjectsInit(session, attributes); err != nil {
		return nil, fmt.Errorf("find public certificate: %w", err)
	}
	// Continue until exhausted; modules may return fewer objects than requested.
	var matches []pkcs11.ObjectHandle
	for len(matches) < 2 {
		objects, _, findErr := module.FindObjects(session, 2-len(matches))
		if findErr != nil {
			_ = module.FindObjectsFinal(session)
			return nil, fmt.Errorf("find public certificate: %w", findErr)
		}
		if len(objects) == 0 {
			break
		}
		matches = append(matches, objects...)
	}
	if err := module.FindObjectsFinal(session); err != nil {
		return nil, fmt.Errorf("finish certificate search: %w", err)
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("expected one public decryption certificate, found %d", len(matches))
	}
	values, err := module.GetAttributeValue(session, matches[0], []*pkcs11.Attribute{pkcs11.NewAttribute(pkcs11.CKA_VALUE, nil)})
	if err != nil {
		return nil, fmt.Errorf("read public certificate: %w", err)
	}
	if len(values) != 1 || values[0] == nil || values[0].Type != pkcs11.CKA_VALUE {
		return nil, fmt.Errorf("public certificate value is missing")
	}
	return x509.ParseCertificate(values[0].Value)
}
