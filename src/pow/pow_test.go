package pow

import (
	"fmt"
	"testing"
)

func TestShift(t *testing.T) {
	pow := NewProofOfWorkT(24)
	fmt.Printf("target zeros: %d\n", (int(256)-pow.Target.BitLen()+1)/8)
	fmt.Printf("target: %v\n", pow.Target.Bytes())

	prevhash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	txHash := []byte{9, 8, 7, 6, 5, 4, 3, 2, 1}

	fmt.Println("mining ...")

	res := pow.solveHash(prevhash, txHash, nil)
	if res == false {
		t.Failed()
	}
	fmt.Println("hash: ", pow.Hash)
	fmt.Println("nonce: ", pow.Nonce)
	fmt.Println("dt:", pow.Duration)
}
