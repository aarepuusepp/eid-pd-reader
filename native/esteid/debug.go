package esteid

import (
	"fmt"
	"os"

	"github.com/ebfe/scard"

	"ee.aarepuusepp.eid-pd-reader/native/apdu"
)

// DebugCardStructure attempts to identify the card structure by trying various SELECT commands
func DebugCardStructure() error {
	fmt.Fprintf(os.Stderr, "\n=== Card Structure Debug Mode ===\n\n")

	ctx, err := scard.EstablishContext()
	if err != nil {
		return fmt.Errorf("pcsc: establish context: %w", err)
	}
	defer ctx.Release()

	rdrs, err := ctx.ListReaders()
	if err != nil {
		return fmt.Errorf("pcsc: list readers: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Found %d reader(s):\n", len(rdrs))
	for i, r := range rdrs {
		fmt.Fprintf(os.Stderr, "  %d: %s\n", i+1, r)
	}

	if len(rdrs) == 0 {
		return fmt.Errorf("no readers found")
	}

	var card *scard.Card
	var connectedReader string
	for _, r := range rdrs {
		c, err := ctx.Connect(r, scard.ShareShared, scard.ProtocolAny)
		if err == nil {
			card = c
			connectedReader = r
			break
		}
	}
	if card == nil {
		return fmt.Errorf("could not connect to any reader")
	}
	defer card.Disconnect(scard.LeaveCard)

	fmt.Fprintf(os.Stderr, "\nConnected to: %s\n", connectedReader)

	// Get ATR
	status, err := card.Status()
	if err == nil {
		fmt.Fprintf(os.Stderr, "ATR: % X\n", status.Atr)
		fmt.Fprintf(os.Stderr, "Protocol: %d\n", status.ActiveProtocol)
	}

	fmt.Fprintf(os.Stderr, "\n--- Testing SELECT MF ---\n")
	if err := debugSelect(card, "MF", [2]byte{0x3F, 0x00}, 0x00); err != nil {
		fmt.Fprintf(os.Stderr, "  FAILED: %v\n", err)
		return err
	}

	// Try selecting EstEID application by AID
	fmt.Fprintf(os.Stderr, "\n--- Testing SELECT by AID ---\n")
	// EstEID AID: A000000077010800070000FE00000100
	esteidAID := []byte{0xA0, 0x00, 0x00, 0x00, 0x77, 0x01, 0x08, 0x00, 0x07, 0x00, 0x00, 0xFE, 0x00, 0x00, 0x01, 0x00}
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x04, P2: 0x00, Data: esteidAID, Le: -1}
	fmt.Fprintf(os.Stderr, "  SELECT EstEID AID: ")
	resp, err := card.Transmit(cmd.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "transmit error: %v\n", err)
	} else {
		sw1, sw2 := apdu.StatusWord(resp)
		fmt.Fprintf(os.Stderr, "SW=%02X%02X", sw1, sw2)
		if sw1 == 0x90 && sw2 == 0x00 {
			fmt.Fprintf(os.Stderr, " ✓ (EstEID application selected)\n")
		} else {
			fmt.Fprintf(os.Stderr, " ✗\n")
		}
	}

	// Try different DFs
	testDFs := []struct {
		name string
		fid  [2]byte
	}{
		{"DF Personal (5000)", [2]byte{0x50, 0x00}},
		{"DF A000", [2]byte{0xA0, 0x00}},
		{"DF EEEE (EstEID)", [2]byte{0xEE, 0xEE}},
		{"DF 0013", [2]byte{0x00, 0x13}},
	}

	fmt.Fprintf(os.Stderr, "\n--- Testing DF Selection ---\n")
	var successDF [2]byte
	var foundDF bool
	for _, df := range testDFs {
		fmt.Fprintf(os.Stderr, "\nTrying %s:\n", df.name)
		
		// Try P1=0x00 (select by FID)
		if err := debugSelect(card, df.name+" (P1=00)", df.fid, 0x00); err == nil {
			successDF = df.fid
			foundDF = true
			break
		}
		
		// Try P1=0x01 (select child DF)
		if err := debugSelect(card, df.name+" (P1=01)", df.fid, 0x01); err == nil {
			successDF = df.fid
			foundDF = true
			break
		}
		
		// Try P1=0x04 (select by DF name)
		if err := debugSelect(card, df.name+" (P1=04)", df.fid, 0x04); err == nil {
			successDF = df.fid
			foundDF = true
			break
		}
	}

	if !foundDF {
		fmt.Fprintf(os.Stderr, "\nNo DF found, trying EFs directly from MF...\n")
		// Re-select MF
		debugSelect(card, "MF (retry)", [2]byte{0x3F, 0x00}, 0x00)
	} else {
		fmt.Fprintf(os.Stderr, "\nSuccessfully selected DF: %02X%02X\n", successDF[0], successDF[1])
	}

	// Try reading various EFs
	testEFs := []struct {
		name string
		fid  [2]byte
	}{
		{"EF 5001 (Surname)", [2]byte{0x50, 0x01}},
		{"EF 5002 (Given Name)", [2]byte{0x50, 0x02}},
		{"EF 5006 (Personal Code)", [2]byte{0x50, 0x06}},
		{"EF 5044 (PD Record File)", [2]byte{0x50, 0x44}},
		{"EF D003 (Document Number)", [2]byte{0xD0, 0x03}},
		{"EF 0016 (Retry Counters)", [2]byte{0x00, 0x16}},
	}

	fmt.Fprintf(os.Stderr, "\n--- Testing EF Reading ---\n")
	for _, ef := range testEFs {
		fmt.Fprintf(os.Stderr, "\nTrying %s:\n", ef.name)
		
		// Try P1=0x02 (select EF by ID)
		cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x02, P2: 0x0C, Data: []byte{ef.fid[0], ef.fid[1]}, Le: -1}
		resp, err := card.Transmit(cmd.Bytes())
		if err != nil {
			fmt.Fprintf(os.Stderr, "  SELECT EF (P1=02): transmit failed: %v\n", err)
			continue
		}
		sw1, sw2 := apdu.StatusWord(resp)
		fmt.Fprintf(os.Stderr, "  SELECT EF (P1=02): SW=%02X%02X", sw1, sw2)
		if sw1 == 0x90 && sw2 == 0x00 {
			fmt.Fprintf(os.Stderr, " ✓\n")
			
			// Try READ BINARY
			readCmd := apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: 256}
			readResp, err := card.Transmit(readCmd.Bytes())
			if err != nil {
				fmt.Fprintf(os.Stderr, "  READ BINARY failed: %v\n", err)
				continue
			}
			rsw1, rsw2 := apdu.StatusWord(readResp)
			fmt.Fprintf(os.Stderr, "  READ BINARY: SW=%02X%02X", rsw1, rsw2)
			if rsw1 == 0x90 && rsw2 == 0x00 {
				data := apdu.DataPart(readResp)
				fmt.Fprintf(os.Stderr, " ✓ (got %d bytes)\n", len(data))
				if len(data) > 0 && len(data) <= 100 {
					fmt.Fprintf(os.Stderr, "  Data: %s\n", sanitizeText(data))
					fmt.Fprintf(os.Stderr, "  Hex: % X\n", data)
				}
			} else {
				fmt.Fprintf(os.Stderr, " ✗\n")
			}
		} else {
			fmt.Fprintf(os.Stderr, " ✗\n")
			
			// Try P1=0x00 (select by FID/path)
			cmd2 := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: 0x00, P2: 0x0C, Data: []byte{ef.fid[0], ef.fid[1]}, Le: -1}
			resp2, err2 := card.Transmit(cmd2.Bytes())
			if err2 != nil {
				fmt.Fprintf(os.Stderr, "  SELECT EF (P1=00): transmit failed: %v\n", err2)
				continue
			}
			sw1_2, sw2_2 := apdu.StatusWord(resp2)
			fmt.Fprintf(os.Stderr, "  SELECT EF (P1=00): SW=%02X%02X", sw1_2, sw2_2)
			if sw1_2 == 0x90 && sw2_2 == 0x00 {
				fmt.Fprintf(os.Stderr, " ✓\n")
				
				// Try READ BINARY
				readCmd := apdu.APDU{CLA: 0x00, INS: 0xB0, P1: 0x00, P2: 0x00, Data: nil, Le: 256}
				readResp, err := card.Transmit(readCmd.Bytes())
				if err != nil {
					fmt.Fprintf(os.Stderr, "  READ BINARY failed: %v\n", err)
					continue
				}
				rsw1, rsw2 := apdu.StatusWord(readResp)
				fmt.Fprintf(os.Stderr, "  READ BINARY: SW=%02X%02X", rsw1, rsw2)
				if rsw1 == 0x90 && rsw2 == 0x00 {
					data := apdu.DataPart(readResp)
					fmt.Fprintf(os.Stderr, " ✓ (got %d bytes)\n", len(data))
					if len(data) > 0 && len(data) <= 100 {
						fmt.Fprintf(os.Stderr, "  Data: %s\n", sanitizeText(data))
						fmt.Fprintf(os.Stderr, "  Hex: % X\n", data)
					}
				} else {
					fmt.Fprintf(os.Stderr, " ✗\n")
				}
			} else {
				fmt.Fprintf(os.Stderr, " ✗\n")
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\n=== Debug Complete ===\n\n")
	return nil
}

func debugSelect(card *scard.Card, name string, fid [2]byte, p1 byte) error {
	cmd := apdu.APDU{CLA: 0x00, INS: 0xA4, P1: p1, P2: 0x0C, Data: []byte{fid[0], fid[1]}, Le: -1}
	fmt.Fprintf(os.Stderr, "  SELECT %s (FID=%02X%02X, P1=%02X): ", name, fid[0], fid[1], p1)
	resp, err := card.Transmit(cmd.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "transmit error: %v\n", err)
		return err
	}
	sw1, sw2 := apdu.StatusWord(resp)
	fmt.Fprintf(os.Stderr, "SW=%02X%02X", sw1, sw2)
	if sw1 == 0x90 && sw2 == 0x00 {
		fmt.Fprintf(os.Stderr, " ✓\n")
		return nil
	}
	fmt.Fprintf(os.Stderr, " ✗\n")
	return fmt.Errorf("SW=%02X%02X", sw1, sw2)
}

