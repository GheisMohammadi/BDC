package block

import (
	"fmt"
	"os"
	"path/filepath"
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
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// exPath := filepath.Dir(dir)
	messagefile, _ := filepath.Abs(dir + "/../../config/block_message.json")

	msg, err := ReadBlockMessage(85, messagefile)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("found msg: ", msg)
}
