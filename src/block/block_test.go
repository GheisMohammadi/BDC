package block

import (
	"fmt"
	"testing"
)

func TestHashBlock(t *testing.T) {
	blk := &Block{
		Header: BlockHeader{
			Timestamp: 42,
		},
		TxsCount:     0,
		Transactions: nil,
		Height:       1,
	}
	hash := blk.GetHash()
	fmt.Println(hash)
	if len(hash) != 32 {
		t.Fatal("Hashing block failed.")
	}
}

func TestMessage(t *testing.T) {
	msg,err := ReadBlockMessage(13)
	if err!=nil {
		t.Error(err)
	}
	fmt.Println("found msg: ",msg)
}