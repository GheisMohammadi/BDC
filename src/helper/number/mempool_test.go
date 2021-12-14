package number

import (
	"fmt"
	"testing"
)

func TestNumber(t *testing.T) {
	ba := Int64ToByteArray(123)
	fmt.Println(ba)
}
