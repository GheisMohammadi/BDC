package block

import (
	"badcoin/src/transaction"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	hash "badcoin/src/helper/hash"
)

const genesisReward = 100
const genesisBlockHeight = 1

type Block struct {
	PrevHash     hash.Hash
	Transactions []transaction.Transaction
	MerkleRoot   hash.Hash
	Nonce        []byte
	Height       uint64
	Timestamp    uint64
	Solution     string
	Difficulty   uint32
}

func (b *Block) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Block %x:", b.GetHash()))

	lines = append(lines, fmt.Sprintf("    PrevBlockHash:   %x", b.PrevHash))
	lines = append(lines, fmt.Sprintf("    Hash:            %x", b.GetHash()))
	lines = append(lines, fmt.Sprintf("    MerkleRoot:      %x", b.MerkleRoot))
	lines = append(lines, fmt.Sprintf("    Timestamp:       %d", b.Timestamp))
	lines = append(lines, fmt.Sprintf("    Difficulty:       %d", b.Difficulty))
	lines = append(lines, fmt.Sprintf("    Nonce:           %d", b.Nonce))
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

func DeserializeBlock(buf []byte) (*Block, error) {
	var blk Block
	err := json.Unmarshal(buf, &blk)
	if err != nil {
		return nil, err
	}
	return &blk, nil
}

func (b *Block) CalcHash() hash.Hash {
	return sha256.Sum256(b.Serialize())
}

func (b *Block) GetHash() hash.Hash {
	return b.CalcHash()
}

func (b *Block) GetPrevHash() hash.Hash {
	return b.PrevHash
}

// SetHeight sets the height of the block
func (b *Block) SetHeight(height uint64) {
	b.Height = height
}
