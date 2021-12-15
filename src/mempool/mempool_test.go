package mempool

import (
	"badcoin/src/transaction"
	"fmt"
	"testing"
	"badcoin/src/wallet"
)

func TestMempool(t *testing.T) {
	mp := NewMempool()
	wal := wallet.NewWallet()
	trans1 := transaction.NewTransaction(wal.PublicKey,0,"receiver1",100,"test data")
	mp.AddTx(trans1)
	if mp.TransactionsCount() != 1 {
		t.Error("adding tx failed")
	}
	mp.Clear()
	fmt.Println(mp.TransactionsCount())
}
