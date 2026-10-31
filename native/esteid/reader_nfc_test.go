package esteid

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ebfe/scard"
)

type transactionCard struct {
	locked           bool
	beginErr, endErr error
	resets           int
}

func (c *transactionCard) BeginTransaction() error {
	if c.beginErr != nil {
		return c.beginErr
	}
	c.locked = true
	return nil
}
func (c *transactionCard) EndTransaction(disposition scard.Disposition) error {
	if !c.locked {
		panic("ending unlocked transaction")
	}
	if disposition != scard.UnpowerCard {
		panic("PACE session was not reset")
	}
	c.resets++
	c.locked = false
	return c.endErr
}

func TestNFCTransactionProtectsReadAndResetsOnFailure(t *testing.T) {
	readErr := errors.New("PACE mapping failed")
	for _, failure := range []error{nil, readErr} {
		card := &transactionCard{}
		want := &PersonalData{Source: "nfc:PACE+SM"}
		pd, err := withNFCCardTransaction(card, func() (*PersonalData, error) {
			if !card.locked {
				t.Fatal("card APDUs not protected from concurrent clients")
			}
			return want, failure
		})
		if !errors.Is(err, failure) || pd != want {
			t.Fatalf("result = %v, %v", pd, err)
		}
		if card.locked || card.resets != 1 {
			t.Fatal("card state leaked after read")
		}
	}
}

func TestNFCTransactionDoesNotReadWithoutLock(t *testing.T) {
	failure := errors.New("card reset by another client")
	card := &transactionCard{beginErr: failure}
	_, err := withNFCCardTransaction(card, func() (*PersonalData, error) {
		t.Fatal("read ran despite lock failure")
		return nil, nil
	})
	if !errors.Is(err, failure) || card.resets != 0 {
		t.Fatal("unexpected cleanup or error")
	}
}

func TestNFCTransactionCleanupFailureDiscardsResult(t *testing.T) {
	failure := errors.New("card removed")
	card := &transactionCard{endErr: failure}
	pd, err := withNFCCardTransaction(card, func() (*PersonalData, error) { return &PersonalData{}, nil })
	if pd != nil || !errors.Is(err, failure) {
		t.Fatal("cleanup failure reported as success")
	}
}

func TestNFCTransactionResetsAfterPanic(t *testing.T) {
	card := &transactionCard{}
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
		if card.locked || card.resets != 1 {
			t.Error("panic leaked card transaction")
		}
	}()
	_, _ = withNFCCardTransaction(card, func() (*PersonalData, error) { panic("read panic") })
}

func TestNFCRetriesOnlyConnectionFailures(t *testing.T) {
	for _, tc := range []struct {
		name    string
		failure error
		calls   int
	}{
		{"power on", &nfcConnectionError{scard.ErrNotReady}, 3},
		{"wrong CAN", errors.New("PACE authentication SW=6300"), 1},
		{"mapping", errors.New("PACE mapping SW=6999"), 1},
		{"secure messaging", errors.New("MAC verification failed"), 1},
		{"cleanup", errors.New("NFC transaction cleanup failed"), 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, sleeps := 0, 0
			_, err := readNFCWithRetry("123456", 3, func(string, int) (*PersonalData, error) { calls++; return nil, tc.failure }, func(time.Duration) { sleeps++ })
			if !errors.Is(err, tc.failure) || calls != tc.calls || sleeps != tc.calls-1 {
				t.Fatalf("err=%v calls=%d sleeps=%d", err, calls, sleeps)
			}
		})
	}
}
func TestNFCRetryStopsAfterConnectionRecovery(t *testing.T) {
	calls := 0
	result := &PersonalData{Source: "nfc:PACE+SM"}
	got, err := readNFCWithRetry("123456", 3, func(string, int) (*PersonalData, error) {
		calls++
		if calls == 1 {
			return nil, &nfcConnectionError{scard.ErrNotReady}
		}
		return result, nil
	}, func(time.Duration) {})
	if err != nil || got != result || calls != 2 {
		t.Fatalf("result=%v err=%v calls=%d", got, err, calls)
	}
}
func TestNFCInvalidCANDoesNotTouchCard(t *testing.T) {
	for _, can := range []string{"", "12345", "12345x", "1234567"} {
		_, err := readNFCWithRetry(can, 3, func(string, int) (*PersonalData, error) { t.Fatal("invalid CAN touched card"); return nil, nil }, func(time.Duration) { t.Fatal("invalid CAN retried") })
		if err == nil {
			t.Fatal("invalid CAN accepted")
		}
	}
}

type personalDataSM struct {
	t                               *testing.T
	thales, empty, missingDirectory bool
	exchanges, reads                int
	fid                             byte
}

func (s *personalDataSM) Transmit(apdu []byte) ([]byte, uint16, error) {
	s.exchanges++
	switch apdu[1] {
	case 0xCA:
		s.t.Fatal("unnecessary GET DATA probe")
	case 0xA4:
		if apdu[2] == 2 {
			s.fid = apdu[6]
			return nil, 0x9000, nil
		}
		if s.missingDirectory || (s.thales && apdu[2] == 1) {
			return nil, 0x6A82, nil
		}
		return nil, 0x9000, nil
	case 0xB0:
		s.reads++
		if s.empty {
			return nil, 0x9000, nil
		}
		return []byte(fmt.Sprintf("field-%d", s.fid)), 0x9000, nil
	}
	return nil, 0x6D00, nil
}
func TestNFCPersonalDataReadsNineFilesWithoutProbes(t *testing.T) {
	for _, thales := range []bool{false, true} {
		for _, empty := range []bool{false, true} {
			sm := &personalDataSM{t: t, thales: thales, empty: empty}
			pd, err := readNFCPersonalData(sm)
			want := 19
			if thales {
				want = 21
			}
			if err != nil || pd == nil || sm.exchanges != want || sm.reads != 9 {
				t.Fatalf("thales=%t empty=%t err=%v exchanges=%d reads=%d", thales, empty, err, sm.exchanges, sm.reads)
			}
			if !empty && (pd.PersonalCode != "field-6" || pd.DocumentNr != "field-7" || pd.IssueDate != "field-9") {
				t.Fatalf("incomplete fields: %+v", pd)
			}
		}
	}
}
func TestNFCDoesNotReadOverlappingFilesOutsidePersonalDataDirectory(t *testing.T) {
	sm := &personalDataSM{t: t, missingDirectory: true}
	pd, err := readNFCPersonalData(sm)
	if err == nil || pd != nil || sm.reads != 0 {
		t.Fatalf("result=%v err=%v reads=%d", pd, err, sm.reads)
	}
}
