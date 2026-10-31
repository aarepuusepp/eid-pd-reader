package nfc

import (
	"errors"
	"github.com/ebfe/scard"
	"testing"
	"time"
)

type fakeReaderContext struct {
	present                  bool
	statusCalls, connections int
	connectErr               error
	card                     scard.Card
	t                        *testing.T
}

func (c *fakeReaderContext) GetStatusChange(states []scard.ReaderState, timeout time.Duration) error {
	c.statusCalls++
	if c.statusCalls > 1 {
		if states[0].CurrentState&scard.StateChanged != 0 {
			c.t.Fatal("StateChanged echoed back to PC/SC")
		}
		if timeout != 5*time.Second {
			c.t.Fatal("missing bounded insertion wait")
		}
		c.present = true
	}
	for i := range states {
		states[i].EventState = scard.StateChanged | scard.StateEmpty
		if c.present {
			states[i].EventState = scard.StateChanged | scard.StatePresent
		}
	}
	return nil
}
func (c *fakeReaderContext) Connect(string, scard.ShareMode, scard.Protocol) (*scard.Card, error) {
	c.connections++
	if c.connectErr != nil {
		return nil, c.connectErr
	}
	return &c.card, nil
}
func TestPresentCardConnectsExactlyOnce(t *testing.T) {
	for _, failure := range []error{nil, scard.ErrNotReady} {
		c := &fakeReaderContext{t: t, present: true, connectErr: failure}
		card, _, err := waitAndConnect(c, []string{"ACS ACR1252 CL Reader PICC 0"})
		if !errors.Is(err, failure) || c.connections != 1 || c.statusCalls != 1 {
			t.Fatalf("card=%v err=%v connections=%d status=%d", card, err, c.connections, c.statusCalls)
		}
	}
}
func TestCardInsertionWaitDoesNotRepeatConnect(t *testing.T) {
	c := &fakeReaderContext{t: t}
	_, _, err := waitAndConnect(c, []string{"ACS ACR1252 CL Reader PICC 0"})
	if err != nil || c.connections != 1 || c.statusCalls != 2 {
		t.Fatalf("err=%v connections=%d status=%d", err, c.connections, c.statusCalls)
	}
}
func TestNFCReaderSelectionExcludesSAM(t *testing.T) {
	for name, want := range map[string]bool{
		"ACS ACR1252 1S CL Reader PICC 0":     true,
		"ACS ACR1252 1S CL Reader SAM 0":      false,
		"Contactless NFC reader":              true,
		"Alcor Micro USB Smart Card Reader 0": false,
	} {
		if got := IsContactlessReader(name); got != want {
			t.Errorf("%q: %t, want %t", name, got, want)
		}
	}
}
