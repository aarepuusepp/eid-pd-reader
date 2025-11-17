package apdu

import (
	"fmt"
)

// APDU is a simple container for command APDUs
// CLA INS P1 P2 Lc Data Le (Le optional)
type APDU struct {
	CLA  byte
	INS  byte
	P1   byte
	P2   byte
	Data []byte
	Le   int // -1 => no Le
}

func (a APDU) Bytes() []byte {
	b := []byte{a.CLA, a.INS, a.P1, a.P2}
	if len(a.Data) > 0 {
		b = append(b, byte(len(a.Data)))
		b = append(b, a.Data...)
	}
	if a.Le >= 0 {
		b = append(b, byte(a.Le))
	}
	return b
}

// StatusWord splits trailer into SW1/SW2
func StatusWord(resp []byte) (byte, byte) {
	if len(resp) < 2 {
		return 0x6F, 0x00 // generic error
	}
	return resp[len(resp)-2], resp[len(resp)-1]
}

// DataPart returns response data without SW1SW2
func DataPart(resp []byte) []byte {
	if len(resp) < 2 {
		return nil
	}
	return resp[:len(resp)-2]
}

func CheckOK(sw1, sw2 byte) error {
	if sw1 == 0x90 && sw2 == 0x00 {
		return nil
	}
	return fmt.Errorf("card returned error SW=%02X%02X", sw1, sw2)
}

