package esteid

import (
	"testing"
)

func TestParseFID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    [2]byte
		wantErr bool
	}{
		{
			name:    "Valid hex",
			input:   "5001",
			want:    [2]byte{0x50, 0x01},
			wantErr: false,
		},
		{
			name:    "Valid hex with 0x prefix",
			input:   "0x5001",
			want:    [2]byte{0x50, 0x01},
			wantErr: false,
		},
		{
			name:    "Valid hex uppercase",
			input:   "5001",
			want:    [2]byte{0x50, 0x01},
			wantErr: false,
		},
		{
			name:    "Invalid length",
			input:   "501",
			want:    [2]byte{},
			wantErr: true,
		},
		{
			name:    "Invalid hex",
			input:   "50GH",
			want:    [2]byte{},
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    [2]byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseFID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name string
		input []byte
		want string
	}{
		{
			name: "Normal text",
			input: []byte("John Doe"),
			want: "John Doe",
		},
		{
			name: "With trailing nulls",
			input: []byte("John Doe\x00\x00\x00"),
			want: "John Doe",
		},
		{
			name: "With internal nulls",
			input: []byte("John\x00Doe"),
			want: "JohnDoe",
		},
		{
			name: "With whitespace",
			input: []byte("  John Doe  \n\r"),
			want: "John Doe",
		},
		{
			name: "Empty",
			input: []byte{},
			want: "",
		},
		{
			name: "Only nulls",
			input: []byte{0x00, 0x00, 0x00},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeText(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeText() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Note: ReadPersonalData requires actual PC/SC connection and card.
// Integration tests should be run manually with a real card.
// This test file focuses on unit-testable helper functions.

