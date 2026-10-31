// eid-pd-card-test runs an explicit local NFC crypto acceptance test.
// It does not use a browser, upload data, persist CAN/PIN, or write plaintext files.
package main

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ee.aarepuusepp.eid-pd-reader/native/esteid"
	"github.com/aarepuusepp/libcdoc-go/v2/cdoc"
)

func main() {
	container := flag.String("container", "", "Local CDOC container; omit for certificate-only probe")
	canEnv := flag.String("can-env", "EIDPD_TEST_CAN", "Environment variable containing CAN (never pass CAN on the command line)")
	flag.Parse()
	if err := run(*container, *canEnv); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}

func run(path, canEnv string) error {
	can := os.Getenv(canEnv)
	_ = os.Unsetenv(canEnv) // Do not pass CAN to the PIN helper process.
	if len(can) != 6 {
		return fmt.Errorf("set the CAN environment variable before running")
	}
	var data []byte
	var err error
	if path != "" {
		data, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}
	fmt.Println("START NFC certificate read (no PIN)")
	result, err := esteid.GetDecryptCertificateNFC(can)
	if err != nil {
		return err
	}
	block, _ := pem.Decode([]byte(result.CertificatePEM))
	if block == nil {
		return fmt.Errorf("invalid certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	fmt.Println("PASS NFC certificate read")
	if path == "" {
		return nil
	}
	challenge, err := cdoc.PrepareClientUnwrap(data, cert)
	if err != nil {
		return fmt.Errorf("prepare container: %w", err)
	}
	raw, err := json.Marshal(challenge)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	fields["displayContext"], err = json.Marshal(map[string]string{"fileName": filepath.Base(path)})
	if err != nil {
		return err
	}
	raw, err = json.Marshal(fields)
	if err != nil {
		return err
	}
	fmt.Println("START NFC key operation; enter PIN1 only in the native dialog")
	unwrapped, err := esteid.UnwrapTransportKeyNFC(can, raw, result.CertificatePEM, "Local EID-PD acceptance test")
	if err != nil {
		return err
	}
	key, err := base64.StdEncoding.DecodeString(unwrapped.ContentEncryptionKeyBase64)
	if err != nil {
		return err
	}
	defer clear(key)
	files, err := cdoc.FinishClientDecrypt(data, cert, key)
	if err != nil {
		return fmt.Errorf("decrypt container: %w", err)
	}
	fmt.Printf("PASS NFC decryption: %d files (plaintext not saved)\n", len(files))
	for i, f := range files {
		fmt.Printf("file %d: bytes=%d sha256=%x\n", i+1, len(f.Content), sha256.Sum256(f.Content))
		clear(f.Content)
	}
	return nil
}
