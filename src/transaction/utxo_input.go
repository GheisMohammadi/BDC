package transaction

import (
	hash "badcoin/src/helper/hash"
	wallet "badcoin/src/wallet"
	"bytes"
	"strconv"
)

// OutPoint defines a badcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash  hash.Hash
	Index int
}

func (o OutPoint) StringHash() string {
	return o.Hash.String()
}

// NewOutPoint returns a new badcoin transaction outpoint point with the
// provided hash and index.
func NewOutPoint(hash *hash.Hash, index int) *OutPoint {
	return &OutPoint{
		Hash:  *hash,
		Index: index,
	}
}

// String returns the OutPoint in the human-readable form "hash:index".
func (o OutPoint) String() string {
	buf := make([]byte, 2*hash.HashSize+1, 2*hash.HashSize+1+10)
	copy(buf, o.Hash.String())
	buf[2*hash.HashSize] = ':'
	buf = strconv.AppendUint(buf, uint64(o.Index), 10)
	return string(buf)
}

// TXInput represents a transaction input
type TXInput struct {
	PreviousOutPoint OutPoint
	Signature        []byte
	PubKey           []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UnLock(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPublicKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// NewTXInput create a new TXInput
func NewTXInput(prevOut *OutPoint, sign, pubKey []byte) *TXInput {
	return &TXInput{*prevOut, sign, pubKey}
}
