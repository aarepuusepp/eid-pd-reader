//go:build !windows

package esteid

func libraryLoads(path string) bool { return false }
