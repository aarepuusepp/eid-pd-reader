package esteid

import "testing"

func TestInstallerPKCS11Decisions(t *testing.T) {
	for _, tc := range []struct {
		state      string
		configured bool
		want       int
	}{
		{"ready", false, 0}, {"ready", true, 0},
		{"missing", false, 10}, {"wrong_architecture", false, 11},
		{"invalid", false, 12}, {"missing", true, 20},
		{"wrong_architecture", true, 20}, {"invalid", true, 20},
	} {
		if got := pkcs11ExitCode(tc.state, tc.configured); got != tc.want {
			t.Errorf("state=%s configured=%v: got %d want %d", tc.state, tc.configured, got, tc.want)
		}
	}
	path := t.TempDir() + "/missing.dll"
	gotPath, state, code := CheckPKCS11(path)
	if gotPath != path || state != "missing" || code != 20 {
		t.Fatalf("explicit missing path: %q %s %d", gotPath, state, code)
	}
}
