//go:build !cgo

package esteid

import (
	"crypto/x509"
	"fmt"

	"github.com/aarepuusepp/libcdoc-go/v2/key"
)

func readPublicCertificate(key.PKCS11Config) (*x509.Certificate, error) {
	return nil, fmt.Errorf("PKCS#11 certificate reading requires a CGO-enabled build")
}
