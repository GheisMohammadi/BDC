package blockchain

import (
	"fmt"
	"math/big"
	"testing"

	config "badcoin/src/config"
	"badcoin/src/transaction"
	"badcoin/src/wallet"

	libp2p "github.com/libp2p/go-libp2p"
)

func TestBigInt(t *testing.T) {
	bal := new(big.Int)
	bal.SetInt64(1000)
	num := new(big.Int)
	num.SetInt64(-1500)
	res := new(big.Int).Add(bal, num)
	fmt.Println(res)
}
func TestAccount(t *testing.T) {

	configs, _ := config.Init("")
	h, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		panic(err)
	}
	bc := NewBlockchain(h, configs)
	bal := new(big.Int)
	bal.SetInt64(1000)
	addr := "asdfhdsjkfbmbhmfvbmxdgjhghsdjfhadgvxaydg"
	err = bc.StoreAccount(&Account{
		Nonce:   13,
		Address: addr,
		Balance: *bal,
	})
	if err != nil {
		t.Error(err)
	}
	err = bc.AddToAccountBalance(addr, 1500)
	if err != nil {
		t.Error(err)
	}
	newbal, _ := bc.GetAccountBalance(addr)
	fmt.Println("Balance: ", newbal.String())
}

func TestUpdateAccounts(t *testing.T) {
	wal1 := wallet.NewWallet()
	wal2 := wallet.NewWallet()
	wal3 := wallet.NewWallet()

	p1 := wal1.PublicKey
	p2 := wal2.PublicKey
	p3 := wal3.PublicKey

	addr1 := wal1.GetStringAddress()
	addr2 := wal2.GetStringAddress()
	addr3 := wal3.GetStringAddress()

	tx1 := transaction.NewTransaction(p1,addr2,300,"1->2")   //acc1: -300    acc2: 300
	tx2 := transaction.NewTransaction(p2,addr3,200,"2->3")   //acc2: 100     acc3: 200 
	tx3 := transaction.NewTransaction(p3,addr1,200,"3->1")   //acc3: 0       acc1: -100
	tx4 := transaction.NewTransaction(p2,addr1,100,"2->1")   //acc2: 0       acc1: 0

	txs := []transaction.Transaction{*tx1,*tx2,*tx3,*tx4}

	configs, _ := config.Init("")
	h, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		panic(err)
	}
	bc := NewBlockchain(h, configs)

	errUpdates := bc.UpdateAccounts(txs)
	if errUpdates!=nil {
		fmt.Println(errUpdates)
		t.Error(errUpdates)
	}
}
