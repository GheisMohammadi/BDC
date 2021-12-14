package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	Init(false)
	Info("test logger")
}
