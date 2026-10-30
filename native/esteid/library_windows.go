package esteid

import "golang.org/x/sys/windows"

func libraryLoads(path string) bool {
	h, err := windows.LoadLibraryEx(path, 0, windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR|windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return false
	}
	defer windows.FreeLibrary(h)
	_, err = windows.GetProcAddress(h, "C_GetFunctionList")
	return err == nil
}
