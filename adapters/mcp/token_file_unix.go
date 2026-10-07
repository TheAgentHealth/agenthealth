//go:build !windows

package mcp

import "os"

func protectTokenFile(file *os.File) error { return file.Chmod(0600) }
func privateTokenFile(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0077 == 0
}
