package utils

import (
	"os"
	"strconv"
)

// ParsePerm parse text `0644` `0755` as os.FileMode
func ParsePerm(permStr string) (os.FileMode, error) {
	var res os.FileMode
	r, err := strconv.ParseInt(permStr, 8, 32)
	if err == nil {
		return os.FileMode(r), nil
	}
	return res, err
}
