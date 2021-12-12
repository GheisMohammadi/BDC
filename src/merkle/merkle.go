package merkle

import (
	"badcoin/src/helper/hash"
	"badcoin/src/transaction"
)

type TxMerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left *MerkleNode, right *MerkleNode, tx *transaction.Transaction) *MerkleNode {
	mnode := MerkleNode{}

	if left == nil && right == nil {
		hash := hash.HashB(tx.Serialize())
		mnode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := hash.HashB(prevHashes)
		mnode.Data = hash[:]
	}
	mnode.Left = left
	mnode.Right = right

	return &mnode
}

func BuildTxMerkleTree(txs []*transaction.Transaction) *TxMerkleTree {
	var nodes []MerkleNode

	if len(txs)%2 != 0 {
		// append nil to txs so that it can be divisible by 2
		txs = append(txs, nil)
	}

	for _, datum := range txs {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(txs)/2; i++ {
		if len(nodes)%2 != 0 {
			nodes = append(nodes, *NewMerkleNode(nil, nil, nil))
		}
		var newLevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	if len(nodes) == 0 {
		zeronode := MerkleNode{
			Left:  nil,
			Right: nil,
			Data:  hash.ZeroHash().CloneBytes(),
		}
		//NewMerkleNode(nil,nil,transaction.NewTransaction([]byte{},"",0,""))
		nodes = append(nodes, zeronode)
	}
	mTree := TxMerkleTree{&nodes[0]}

	return &mTree
}
