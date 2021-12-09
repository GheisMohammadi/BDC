package block 

import (
    "fmt"
    "testing"
)

func TestHashBlock(t* testing.T){
    blk := &Block{
        Transactions: nil,
        Height: 100,
        Timestamp: 42,
    }
    hash := blk.GetHash()
    fmt.Println(hash)
    if len(hash) != 32 {
        t.Fatal("Hashing block failed.")
    }
}
