package uuid

import (
	"fmt"
	"testing"
)

func TestUUID(t *testing.T) {
	newuuid := NewV4()
	fmt.Println(newuuid)
}
