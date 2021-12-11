package address

import (
	wallet "badcoin/src/wallet"
	"fmt"
	"testing"
)

func TestBlockchain(t *testing.T) {
	newwallet := wallet.NewWallet()
	pubKey := newwallet.PublicKey
	addr := FromPublicKey(pubKey)
	addrstring := ToString(addr)
	pubhash := HashPublicKey(pubKey)
	pubhash2 := GetPubKeyHashFromAddress(addr)
	fmt.Printf("pub: %v\n", pubKey)
	fmt.Printf("pub hash: %v\n", pubhash)
	fmt.Printf("address: %s\n", addrstring)
	fmt.Printf("len address: %d\n", len(ToString(addr)))
	fmt.Printf("pub hash: %v\n", pubhash2)
	if string(pubhash) != string(pubhash2) {
		t.Error("pub hash not match")
	}
}
