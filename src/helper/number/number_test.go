package number

import (
	"fmt"
	"math/big"
	"testing"
)

func TestNumber(t *testing.T) {
	ba := Int64ToByteArray(123)
	fmt.Println(ba)

	res := RoundBigFloat(big.NewFloat(34.123456789))
	fmt.Println(res)
}
