package esteid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPKCS11DiagnosticsMissingAndInvalid(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.dll")
	if got := pkcs11State(missing); got != "missing" {
		t.Fatalf("missing DLL: %s", got)
	}
	invalid := filepath.Join(t.TempDir(), "invalid.dll")
	if err := os.WriteFile(invalid, []byte("not a DLL"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := pkcs11State(invalid); got != "invalid" {
		t.Fatalf("invalid DLL: %s", got)
	}
}
