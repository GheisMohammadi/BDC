package merkle

import (
	"badcoin/src/transaction"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNewMerkleNode(t *testing.T) {

	trans1 := transaction.Transaction{From: "sender1", To: "receiver1", Value: 100, Data: "Data1"}
	trans2 := transaction.Transaction{From: "sender2", To: "receiver2", Value: 200, Data: "Data2"}
	trans3 := transaction.Transaction{From: "sender3", To: "receiver3", Value: 300, Data: "Data3"}

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

	if "1b37eba70eba8f9296ac8d4ac5898d69467e4e31366b6cac79dadf5f81579298" == hex.EncodeToString(n5.Data) {
		t.Log(hex.EncodeToString(n5.Data))
	} else {
		t.Error("Level 1 hash 1 is correct", hex.EncodeToString(n5.Data))
	}

	if "10266b901c67ffbf48489d9b88614368ad9c00f5cc65e912cf7bdabe30d5cb06" == hex.EncodeToString(n6.Data) {
		t.Log(hex.EncodeToString(n6.Data))
	} else {
		t.Error("Level 1 hash 2 is correct", hex.EncodeToString(n6.Data))
	}

	if "38addfe1aa839ab8842dcc11b284e027b0c43332fc3bd8b477413075bde5ea18" == hex.EncodeToString(n7.Data) {
		t.Log(hex.EncodeToString(n7.Data))
	} else {
		t.Error("Root hash is correct", hex.EncodeToString(n7.Data))
	}
}

func TestNewMerkleTree(t *testing.T) {

	trans1 := transaction.Transaction{From: "sender1", To: "receiver1", Value: 100, Data: "Data1"}
	trans2 := transaction.Transaction{From: "sender2", To: "receiver2", Value: 200, Data: "Data2"}
	trans3 := transaction.Transaction{From: "sender3", To: "receiver3", Value: 300, Data: "Data3"}
	trans4 := transaction.Transaction{From: "sender4", To: "receiver3", Value: 400, Data: "Data4"}
	trans5 := transaction.Transaction{From: "sender5", To: "receiver5", Value: 500, Data: "Data5"}

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
