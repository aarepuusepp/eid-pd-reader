package policy

import (
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func machineDataDir() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
}
func machinePolicy() ([]byte, bool, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Policies\EIDPD`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err == registry.ErrNotExist {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer k.Close()
	value, _, err := k.GetStringValue("Policy")
	if err != nil {
		return nil, true, fmt.Errorf("machine policy: %w", err)
	}
	return []byte(value), true, nil
}
