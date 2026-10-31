package nfc

import (
	"bytes"
	"testing"
)

type smFunc func([]byte) ([]byte, uint16, error)

func (f smFunc) Transmit(b []byte) ([]byte, uint16, error) { return f(b) }
func TestSMCertificateReadUsesBoundedOffsets(t *testing.T) {
	want := bytes.Repeat([]byte{0xA5}, 1731)
	calls := 0
	sm := smFunc(func(cmd []byte) ([]byte, uint16, error) {
		calls++
		offset := int(cmd[2])<<8 | int(cmd[3])
		length := int(cmd[4])
		if length == 0 || length > 223 {
			t.Fatal("unsafe SM Le")
		}
		end := offset + length
		sw := uint16(0x9000)
		if end > len(want) {
			end = len(want)
			sw = 0x6282
		}
		return want[offset:end], sw, nil
	})
	got, err := SMReadEF(sm, 8192)
	if err != nil || !bytes.Equal(got, want) || calls != 8 {
		t.Fatalf("read: %d calls, %d bytes, %v", calls, len(got), err)
	}
}
func TestSMReadRejectsOversizedResponse(t *testing.T) {
	_, err := SMReadEF(smFunc(func([]byte) ([]byte, uint16, error) { return make([]byte, 224), 0x9000, nil }), 512)
	if err == nil {
		t.Fatal("accepted excessive response")
	}
}
func TestSMReadCorrectsFinalLength(t *testing.T) {
	calls := 0
	got, err := SMReadEF(smFunc(func(cmd []byte) ([]byte, uint16, error) {
		calls++
		if calls == 1 {
			return nil, 0x6C05, nil
		}
		if cmd[4] != 5 {
			t.Fatal("incorrect corrected Le")
		}
		return []byte("final"), 0x9000, nil
	}), 512)
	if err != nil || string(got) != "final" || calls != 2 {
		t.Fatalf("%s %v", got, err)
	}
}
