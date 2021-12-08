package file

import (
	"os"
)

func IsExist(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}

	return true
}
