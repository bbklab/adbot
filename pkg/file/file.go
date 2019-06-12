package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// AtomicWriteFile write bytes to disk file with atomic guarantee
// FIXME bettwer to use the temp file under the same directory as the origin path
// thus to prevent IO load by cross different disk partition or
// errors like: invalid cross-device link
func AtomicWriteFile(path string, data []byte, mode os.FileMode) (err error) {
	f, err := ioutil.TempFile(filepath.Dir(path), ".atomic-temp-")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	n, err := f.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("AtomicWriteFile wrote less than expected")
	}

	if err := f.Sync(); err != nil {
		return err
	}

	f.Close()
	os.Chmod(f.Name(), mode)
	return os.Rename(f.Name(), path) // rename sysCall is atomic under linux
}

// CreateIfNotExists creates a file or a directory only if it does not already exist.
func CreateIfNotExists(path string, isDir bool) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if isDir {
				return os.MkdirAll(path, 0755)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			f.Close()
		}
	}
	return nil
}

// CopyFile copies from src to dst until either EOF is reached
// on src or an error occurs. It verifies src exists and remove
// the dst if it exists.
func CopyFile(src, dst string) (int64, error) {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return 0, nil
	}
	sf, err := os.Open(cleanSrc)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	if err := os.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	df, err := os.Create(cleanDst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

// SubFiles list all of filenames under given directory
func SubFiles(dir string) []string {
	finfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	var ret = make([]string, 0)
	for _, finfo := range finfos {
		if !finfo.Mode().IsRegular() {
			continue
		}
		ret = append(ret, finfo.Name())
	}

	return ret
}

// Sha256File caculate the sha256sum of given file path
func Sha256File(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// IsEmptyFile detect if given file path is an empty file
func IsEmptyFile(path string) bool {
	finfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return finfo.Size() == 0 && finfo.Mode().IsRegular()
}

// Exists detect if given path is an exists file
func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
