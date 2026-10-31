package esteid

import (
	"debug/pe"
	"ee.aarepuusepp.eid-pd-reader/native/nfc"
	"github.com/ebfe/scard"
	"os"
	"strings"
)

// Diagnostics only checks reader/card presence and the library architecture.
// It does not connect to a card, read identifiers/certificates or request a PIN.
func Diagnostics(library string) map[string]interface{} {
	result := map[string]interface{}{"readerCount": 0, "cardPresent": false, "pcscAvailable": false, "pkcs11State": "missing"}
	result["nfcReaderCount"], result["contactReaderCount"] = 0, 0
	result["nfcCardPresent"], result["contactCardPresent"] = false, false
	if library == "" {
		library = DefaultPKCS11LibraryPath()
	}
	result["pkcs11State"] = pkcs11State(library)

	ctx, err := scard.EstablishContext()
	if err != nil {
		return result
	}
	defer ctx.Release()
	result["pcscAvailable"] = true
	readers, err := ctx.ListReaders()
	if err != nil {
		return result
	}
	result["readerCount"] = len(readers)
	states := make([]scard.ReaderState, len(readers))
	for i, name := range readers {
		if !strings.Contains(strings.ToLower(name), "sam") {
			key := "contactReaderCount"
			if nfc.IsContactlessReader(name) {
				key = "nfcReaderCount"
			}
			result[key] = result[key].(int) + 1
		}
		states[i].Reader = name
		states[i].CurrentState = scard.StateUnaware
	}
	if len(states) > 0 && ctx.GetStatusChange(states, 0) == nil {
		for _, state := range states {
			if state.EventState&scard.StatePresent != 0 {
				result["cardPresent"] = true
				if !strings.Contains(strings.ToLower(state.Reader), "sam") {
					key := "contactCardPresent"
					if nfc.IsContactlessReader(state.Reader) {
						key = "nfcCardPresent"
					}
					result[key] = true
				}
			}
		}
	}
	return result
}

func pkcs11State(library string) string {
	if library == "" {
		return "missing"
	}
	if _, err := os.Stat(library); os.IsNotExist(err) {
		return "missing"
	}
	file, err := pe.Open(library)
	if err != nil {
		return "invalid"
	}
	defer file.Close()
	if file.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
		return "wrong_architecture"
	}
	if !libraryLoads(library) {
		return "invalid"
	}
	return "ready"
}
