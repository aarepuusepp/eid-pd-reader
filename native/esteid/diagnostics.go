package esteid

import (
	"debug/pe"
	"github.com/ebfe/scard"
	"os"
)

// Diagnostics only checks reader/card presence and the library architecture.
// It does not connect to a card, read identifiers/certificates or request a PIN.
func Diagnostics(library string) map[string]interface{} {
	result := map[string]interface{}{"readerCount": 0, "cardPresent": false, "pcscAvailable": false, "pkcs11State": "missing"}
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
		states[i].Reader = name
		states[i].CurrentState = scard.StateUnaware
	}
	if len(states) > 0 && ctx.GetStatusChange(states, 0) == nil {
		for _, state := range states {
			if state.EventState&scard.StatePresent != 0 {
				result["cardPresent"] = true
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
