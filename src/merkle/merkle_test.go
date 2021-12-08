package merkle

import (
	"badcoin/src/transaction"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNewMerkleNode(t *testing.T) {

	trans1 := transaction.Transaction{Sender: "sender1", Receiver: "receiver1", Amount: 100, Memo: "memo1"}
	trans2 := transaction.Transaction{Sender: "sender2", Receiver: "receiver2", Amount: 200, Memo: "memo2"}
	trans3 := transaction.Transaction{Sender: "sender3", Receiver: "receiver3", Amount: 300, Memo: "memo3"}

	// Level 1
	n1 := NewMerkleNode(nil, nil, &trans1)
	n2 := NewMerkleNode(nil, nil, &trans2)
	n3 := NewMerkleNode(nil, nil, &trans3)
	n4 := NewMerkleNode(nil, nil, nil)

	// Level 2
	n5 := NewMerkleNode(n1, n2, nil)
	n6 := NewMerkleNode(n3, n4, nil)

	// Level 3
	n7 := NewMerkleNode(n5, n6, nil)

	if "23d8653ced698540540c3cb321e4305fb51dcda39a0624e0bf8f4b8f61f96391" == hex.EncodeToString(n5.Data) {
		t.Log(hex.EncodeToString(n5.Data))
	} else {
		t.Error("Level 1 hash 1 is correct", hex.EncodeToString(n5.Data))
	}

	if "42042d1daa21f80954e30d9445b3047a84c818b6b9104683c340bdfcca3e778c" == hex.EncodeToString(n6.Data) {
		t.Log(hex.EncodeToString(n6.Data))
	} else {
		t.Error("Level 1 hash 2 is correct", hex.EncodeToString(n6.Data))
	}

	if "668ec5c19a420cfc653861757fd8ad110cc6c3143948680b6e0d46a1671ed8df" == hex.EncodeToString(n7.Data) {
		t.Log(hex.EncodeToString(n7.Data))
	} else {
		t.Error("Root hash is correct", hex.EncodeToString(n7.Data))
	}
}

func TestNewMerkleTree(t *testing.T) {

	trans1 := transaction.Transaction{Sender: "sender1", Receiver: "receiver1", Amount: 100, Memo: "memo1"}
	trans2 := transaction.Transaction{Sender: "sender2", Receiver: "receiver2", Amount: 200, Memo: "memo2"}
	trans3 := transaction.Transaction{Sender: "sender3", Receiver: "receiver3", Amount: 300, Memo: "memo3"}
	trans4 := transaction.Transaction{Sender: "sender4", Receiver: "receiver3", Amount: 400, Memo: "memo4"}
	trans5 := transaction.Transaction{Sender: "sender5", Receiver: "receiver5", Amount: 500, Memo: "memo5"}

	txs := []*transaction.Transaction{&trans1, &trans2, &trans3, &trans4, &trans5}

	// Level 1
	n1 := NewMerkleNode(nil, nil, &trans1)
	n2 := NewMerkleNode(nil, nil, &trans2)
	n3 := NewMerkleNode(nil, nil, &trans3)
	n4 := NewMerkleNode(nil, nil, &trans4)
	n5 := NewMerkleNode(nil, nil, &trans5)
	n6 := NewMerkleNode(nil, nil, nil)

	// for loop i=0
	// Level 2
	n11 := NewMerkleNode(n1, n2, nil)
	n12 := NewMerkleNode(n3, n4, nil)
	n13 := NewMerkleNode(n5, n6, nil)
	n14 := NewMerkleNode(nil, nil, nil)

	// for loop i=1
	// Level 3
	n21 := NewMerkleNode(n11, n12, nil)
	n22 := NewMerkleNode(n13, n14, nil)

	n31 := NewMerkleNode(n21, n22, nil)

	rootHash := fmt.Sprintf("%x", n31.Data)
	mTree := BuildTxMerkleTree(txs)

	if rootHash == fmt.Sprintf("%x", mTree.RootNode.Data) {
		t.Log(rootHash)
		t.Log(fmt.Sprintf("%x", mTree.RootNode.Data))
	} else {
		t.Error("Merkle tree root hash is correct")
	}
}
