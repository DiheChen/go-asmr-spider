package utils

import (
	"errors"
	"os"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || errors.Is(err, os.ErrExist)
}
