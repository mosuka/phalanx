package util

import (
	"errors"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsFile(path string) bool {
	if f, err := os.Stat(path); errors.Is(err, os.ErrNotExist) || f.IsDir() {
		return false
	} else {
		return true
	}
}

func IsDir(path string) bool {
	if f, err := os.Stat(path); errors.Is(err, os.ErrNotExist) || !f.IsDir() {
		return false
	} else {
		return true
	}
}
