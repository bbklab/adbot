package dir

import "os"

// Exists detect if given path is an exists directory
func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsDir()
}
