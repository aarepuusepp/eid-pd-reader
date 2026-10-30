//go:build cgo

package esteid

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/aarepuusepp/libcdoc-go/v2/key"
	"github.com/miekg/pkcs11"
)

type publicModuleFake struct {
	calls       []string
	fail        string
	objects     [][]pkcs11.ObjectHandle
	der         []byte
	attributes  []*pkcs11.Attribute
	slot, flags uint
}

func (m *publicModuleFake) call(name string) error {
	m.calls = append(m.calls, name)
	if m.fail == name {
		return errors.New("test module failure")
	}
	return nil
}
func (m *publicModuleFake) Initialize(...pkcs11.InitializeOption) error { return m.call("initialize") }
func (m *publicModuleFake) Finalize() error                             { return m.call("finalize") }
func (m *publicModuleFake) Destroy()                                    { _ = m.call("destroy") }
func (m *publicModuleFake) GetSlotList(bool) ([]uint, error)            { return []uint{10, 20}, m.call("slots") }
func (m *publicModuleFake) OpenSession(slot, flags uint) (pkcs11.SessionHandle, error) {
	m.slot, m.flags = slot, flags
	return 1, m.call("open")
}
func (m *publicModuleFake) CloseSession(pkcs11.SessionHandle) error { return m.call("close") }
func (m *publicModuleFake) FindObjectsInit(_ pkcs11.SessionHandle, attributes []*pkcs11.Attribute) error {
	m.attributes = attributes
	return m.call("findInit")
}
func (m *publicModuleFake) FindObjects(pkcs11.SessionHandle, int) ([]pkcs11.ObjectHandle, bool, error) {
	err := m.call("find")
	if len(m.objects) == 0 {
		return nil, false, err
	}
	result := m.objects[0]
	m.objects = m.objects[1:]
	return result, false, err
}
func (m *publicModuleFake) FindObjectsFinal(pkcs11.SessionHandle) error { return m.call("findFinal") }
func (m *publicModuleFake) GetAttributeValue(pkcs11.SessionHandle, pkcs11.ObjectHandle, []*pkcs11.Attribute) ([]*pkcs11.Attribute, error) {
	return []*pkcs11.Attribute{pkcs11.NewAttribute(pkcs11.CKA_VALUE, m.der)}, m.call("value")
}

func TestPublicCertificateWithoutLogin(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	m := &publicModuleFake{der: der, objects: [][]pkcs11.ObjectHandle{{42}}}
	cert, err := readPublicCertificateFromModule(m, key.PKCS11Config{Slot: 1, KeyID: []byte{1}, Label: "Authentication"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(cert.Raw, der) {
		t.Fatal("wrong certificate")
	}
	if m.slot != 20 || m.flags != pkcs11.CKF_SERIAL_SESSION {
		t.Fatal("expected selected slot and read-only session")
	}
	expected := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_CERTIFICATE),
		pkcs11.NewAttribute(pkcs11.CKA_CERTIFICATE_TYPE, pkcs11.CKC_X_509),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_ID, []byte{1}),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, "Authentication"),
	}
	if !reflect.DeepEqual(m.attributes, expected) {
		t.Fatal("certificate selector changed")
	}
	want := []string{"initialize", "slots", "open", "findInit", "find", "find", "findFinal", "value", "close", "finalize", "destroy"}
	if !reflect.DeepEqual(m.calls, want) {
		t.Fatalf("calls: %v", m.calls)
	}
}

func TestPublicCertificateRejectsMissingAmbiguousAndInvalid(t *testing.T) {
	for _, tc := range []struct {
		name    string
		objects [][]pkcs11.ObjectHandle
	}{
		{"missing", nil}, {"ambiguous", [][]pkcs11.ObjectHandle{{1}, {2}}}, {"invalid DER", [][]pkcs11.ObjectHandle{{1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &publicModuleFake{objects: tc.objects, der: []byte("invalid")}
			if _, err := readPublicCertificateFromModule(m, key.PKCS11Config{}); err == nil {
				t.Fatal("expected error")
			}
			if !reflect.DeepEqual(m.calls[len(m.calls)-3:], []string{"close", "finalize", "destroy"}) {
				t.Fatalf("cleanup: %v", m.calls)
			}
		})
	}
}

func TestPublicCertificateCleansUpFailures(t *testing.T) {
	for _, tc := range []struct {
		fail string
		tail []string
	}{
		{"initialize", []string{"destroy"}},
		{"slots", []string{"finalize", "destroy"}},
		{"open", []string{"finalize", "destroy"}},
		{"findInit", []string{"close", "finalize", "destroy"}},
		{"find", []string{"findFinal", "close", "finalize", "destroy"}},
		{"findFinal", []string{"close", "finalize", "destroy"}},
		{"value", []string{"close", "finalize", "destroy"}},
	} {
		t.Run(tc.fail, func(t *testing.T) {
			m := &publicModuleFake{fail: tc.fail, objects: [][]pkcs11.ObjectHandle{{1}}}
			if _, err := readPublicCertificateFromModule(m, key.PKCS11Config{}); err == nil {
				t.Fatal("expected module error")
			}
			if !reflect.DeepEqual(m.calls[len(m.calls)-len(tc.tail):], tc.tail) {
				t.Fatalf("cleanup: %v", m.calls)
			}
		})
	}
}

func TestPublicCertificateConfigDoesNotRequirePIN(t *testing.T) {
	cfg, err := ResolvePKCS11Config("test.dll", 0, "", "01", "", nil, "", "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PIN != "" || cfg.LibraryPath != "test.dll" {
		t.Fatal("unexpected public certificate config")
	}
}
