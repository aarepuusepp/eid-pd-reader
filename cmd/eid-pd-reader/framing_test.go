package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"
)

type fragmentedReader struct {
	io.Reader
	chunk int
}

func (r fragmentedReader) Read(p []byte) (int, error) {
	if len(p) > r.chunk {
		p = p[:r.chunk]
	}
	return r.Reader.Read(p)
}

func nativeFrame(payload []byte) []byte {
	frame := make([]byte, 4, 4+len(payload))
	binary.LittleEndian.PutUint32(frame, uint32(len(payload)))
	return append(frame, payload...)
}

func TestReadNativeMessageFragmented(t *testing.T) {
	// Certificate and unwrap requests exceed bufio's default buffer size.
	payload := []byte(`{"command":"unwrap-transport-key","certificate":"` + strings.Repeat("A", 12000) + `"}`)
	for _, chunk := range []int{1, 3, 128, 4096} {
		r := bufio.NewReader(fragmentedReader{bytes.NewReader(append(nativeFrame(payload), nativeFrame([]byte(`{"command":"ping"}`))...)), chunk})
		for _, want := range [][]byte{payload, []byte(`{"command":"ping"}`)} {
			got, err := readNativeMessage(r)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("chunk %d: incomplete or corrupted frame (length %d), error %v", chunk, len(got), err)
			}
		}
	}
}

func TestReadNativeMessageTruncated(t *testing.T) {
	frame := nativeFrame([]byte(`{"command":"ping"}`))
	for _, size := range []int{1, 3, 5, len(frame) - 1} {
		if _, err := readNativeMessage(bufio.NewReader(bytes.NewReader(frame[:size]))); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("size %d: expected truncated frame error, got %v", size, err)
		}
	}
	if _, err := readNativeMessage(bufio.NewReader(bytes.NewReader(nil))); !errors.Is(err, io.EOF) {
		t.Fatalf("empty stream: %v", err)
	}
}

func TestReadNativeMessageRejectsInvalidLength(t *testing.T) {
	for _, size := range []uint32{0, 1<<20 + 1, ^uint32(0)} {
		frame := make([]byte, 4)
		binary.LittleEndian.PutUint32(frame, size)
		if _, err := readNativeMessage(bufio.NewReader(bytes.NewReader(frame))); err == nil {
			t.Fatalf("accepted length %d", size)
		}
	}
}
