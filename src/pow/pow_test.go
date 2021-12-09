package pow

import (
	"fmt"
	"math/big"
	"testing"
)

func TestShift(t *testing.T) {
	pow := NewProofOfWorkT(16)
	fmt.Println(pow.target)
	
	target := big.NewInt(2)
	targetBits:=2
	target.Lsh(target, uint(targetBits))
	fmt.Println(target)
}
