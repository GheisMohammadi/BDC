package block

import (
	"badcoin/src/transaction"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	hash "badcoin/src/helper/hash"
)

type BlockHeader struct {
	Version    uint64
	PrevHash   hash.Hash
	MerkleRoot hash.Hash
	Timestamp  int64
	Nonce      int64
	Miner      string
	Difficulty uint32
}

type Block struct {
	Height       uint64
	Hash         hash.Hash
	Header       BlockHeader
	Reward       *big.Float
	TxsCount     uint64
	Transactions []*transaction.Transaction
}

func (b *Block) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Block %x:", b.GetHash()))

	lines = append(lines, fmt.Sprintf("    PrevBlockHash:   %x", b.Header.PrevHash))
	lines = append(lines, fmt.Sprintf("    Hash:            %x", b.GetHash()))
	lines = append(lines, fmt.Sprintf("    MerkleRoot:      %x", b.Header.MerkleRoot))
	lines = append(lines, fmt.Sprintf("    Timestamp:       %d", b.Header.Timestamp))
	lines = append(lines, fmt.Sprintf("    Difficulty:      %d", b.Header.Difficulty))
	lines = append(lines, fmt.Sprintf("    Nonce:           %d", b.Header.Nonce))
	lines = append(lines, fmt.Sprintf("    Miner:           %s", b.Header.Miner))
	lines = append(lines, fmt.Sprintf("    Height:          %d", b.Height))

	lines = append(lines, fmt.Sprintf("    Transactions     %d:", len(b.Transactions)))
	for _, tx := range b.Transactions {
		lines = append(lines, fmt.Sprintf(tx.String()))
	}

	return strings.Join(lines, "\n")
}

func (b *Block) Serialize() []byte {
	data, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return data
}

func (bh *BlockHeader) Serialize() []byte {
	data, err := json.Marshal(bh)
	if err != nil {
		panic(err)
	}
	return data
}

func DeserializeBlock(buf []byte) (*Block, error) {
	var blk Block
	err := json.Unmarshal(buf, &blk)
	if err != nil {
		return nil, err
	}
	return &blk, nil
}

func (b *Block) CalcHash() hash.Hash {
	return hash.HashH(b.Header.Serialize())
}

func (b *Block) GetHash() hash.Hash {
	return b.CalcHash()
}

func (b *Block) GetHashBytes() []byte {
	h := b.CalcHash()
	return h.CloneBytes()
}

func (b *Block) GetPrevHash() hash.Hash {
	return b.Header.PrevHash
}

// SetHeight sets the height of the block
func (b *Block) SetHeight(height uint64) {
	b.Height = height
}

// SetHash sets the hash of the block
func (b *Block) SetHash(h hash.Hash) {
	b.Hash = h
}

// SetHash sets the hash of the block
func (b *Block) UpdateHash() {
	b.Hash = b.CalcHash()
}
