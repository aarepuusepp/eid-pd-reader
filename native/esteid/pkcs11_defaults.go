package esteid

import (
	"debug/pe"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultPKCS11LibraryPath returns the platform default OpenSC PKCS#11 module path when installed.
func DefaultPKCS11LibraryPath() string {
	switch runtime.GOOS {
	case "windows":
		candidates := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "OpenSC Project", "OpenSC", "pkcs11", "opensc-pkcs11.dll"),
			`C:\Windows\System32\opensc-pkcs11.dll`,
			`C:\Windows\System32\onepin-opensc-pkcs11.dll`,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "OpenSC Project", "OpenSC", "pkcs11", "opensc-pkcs11.dll"),
		}
		fallback := ""
		for _, c := range candidates {
			if _, err := os.Stat(c); err != nil {
				continue
			}
			if fallback == "" {
				fallback = c
			}
			f, err := pe.Open(c)
			if err != nil {
				continue
			}
			machine := f.Machine
			f.Close()
			if machine == pe.IMAGE_FILE_MACHINE_AMD64 && libraryLoads(c) {
				return c
			}
		}
		return fallback
	case "darwin":
		return "/Library/OpenSC/lib/opensc-pkcs11.so"
	default:
		candidates := []string{
			"/usr/lib/x86_64-linux-gnu/opensc-pkcs11.so",
			"/usr/lib/opensc-pkcs11.so",
			"/usr/local/lib/opensc-pkcs11.so",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
		return candidates[0]
	}
}

// DefaultAuthKeyIDHex is the Estonian ID-card authentication key CKA_ID in OpenSC.
const DefaultAuthKeyIDHex = "01"
