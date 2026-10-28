package apdu

import (
	"testing"
)

func TestAPDUBytes(t *testing.T) {
	tests := []struct {
		name string
		apdu APDU
		want []byte
	}{
		{
			name: "SELECT MF",
			apdu: APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{0x3F, 0x00}, Le: -1},
			want: []byte{0x00, 0xA4, 0x00, 0x0C, 0x02, 0x3F, 0x00},
		},
		{
			name: "READ BINARY with Le 256 encoded as zero",
			apdu: APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: 256},
			want: []byte{0x00, 0xB0, 0x00, 0x00, 0x00},
		},
		{
			name: "READ BINARY with Le 255",
			apdu: APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Le: 255},
			want: []byte{0x00, 0xB0, 0x00, 0x00, 0xFF},
		},
		{
			name: "No data, no Le",
			apdu: APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: nil, Le: -1},
			want: []byte{0x00, 0xA4, 0x00, 0x0C},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apdu.Bytes()
			if len(got) != len(tt.want) {
				t.Errorf("Bytes() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Bytes()[%d] = %02X, want %02X", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestStatusWord(t *testing.T) {
	tests := []struct {
		name    string
		resp    []byte
		wantSW1 byte
		wantSW2 byte
	}{
		{
			name:    "Success response",
			resp:    []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x90, 0x00},
			wantSW1: 0x90,
			wantSW2: 0x00,
		},
		{
			name:    "Error response",
			resp:    []byte{0x6A, 0x82},
			wantSW1: 0x6A,
			wantSW2: 0x82,
		},
		{
			name:    "Short response",
			resp:    []byte{0x90},
			wantSW1: 0x6F,
			wantSW2: 0x00,
		},
		{
			name:    "Empty response",
			resp:    []byte{},
			wantSW1: 0x6F,
			wantSW2: 0x00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw1, sw2 := StatusWord(tt.resp)
			if sw1 != tt.wantSW1 || sw2 != tt.wantSW2 {
				t.Errorf("StatusWord() = (%02X, %02X), want (%02X, %02X)", sw1, sw2, tt.wantSW1, tt.wantSW2)
			}
		})
	}
}

func TestDataPart(t *testing.T) {
	tests := []struct {
		name string
		resp []byte
		want []byte
	}{
		{
			name: "With data",
			resp: []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x90, 0x00},
			want: []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name: "Only SW",
			resp: []byte{0x90, 0x00},
			want: []byte{},
		},
		{
			name: "Short response",
			resp: []byte{0x90},
			want: nil,
		},
		{
			name: "Empty response",
			resp: []byte{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DataPart(tt.resp)
			if len(got) != len(tt.want) {
				t.Errorf("DataPart() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("DataPart()[%d] = %02X, want %02X", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCheckOK(t *testing.T) {
	tests := []struct {
		name string
		sw1  byte
		sw2  byte
		want error
	}{
		{
			name: "Success",
			sw1:  0x90,
			sw2:  0x00,
			want: nil,
		},
		{
			name: "File not found",
			sw1:  0x6A,
			sw2:  0x82,
			want: nil, // Will be non-nil error
		},
		{
			name: "Wrong P1P2",
			sw1:  0x6A,
			sw2:  0x86,
			want: nil, // Will be non-nil error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckOK(tt.sw1, tt.sw2)
			if tt.want == nil {
				if err != nil && tt.sw1 == 0x90 && tt.sw2 == 0x00 {
					t.Errorf("CheckOK() = %v, want nil", err)
				}
				if err == nil && (tt.sw1 != 0x90 || tt.sw2 != 0x00) {
					t.Errorf("CheckOK() = nil, want error")
				}
			}
		})
	}
}
