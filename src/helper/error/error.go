package error

import (
	"errors"
)

var BlockNotFount = errors.New("Block is not found")
var StorageInitFailed = errors.New("storage initialization failed")