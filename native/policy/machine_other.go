//go:build !windows

package policy

func machineDataDir() (string, error)      { return "/etc", nil }
func machinePolicy() ([]byte, bool, error) { return nil, false, nil }
