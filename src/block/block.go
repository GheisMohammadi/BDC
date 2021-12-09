package block

import (
	"badcoin/src/transaction"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	mh "github.com/multiformats/go-multihash"
	hash "badcoin/src/helper/hash"
)

const genesisReward = 100
const genesisBlockHeight = 1

type Block struct {
	PrevHash     hash.Hash
	Transactions []transaction.Transaction
	MerkleRoot   hash.Hash
	Nonce        []byte
	Hash         hash.Hash
	Height       uint64
	Timestamp    uint64
	Solution     string
	Difficult    uint32
}

func (b *Block) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Block %x:", b.Hash))

	lines = append(lines, fmt.Sprintf("    PrevBlockHash:   %x", b.PrevHash))
	lines = append(lines, fmt.Sprintf("    Hash:            %x", b.Hash))
	lines = append(lines, fmt.Sprintf("    MerkleRoot:      %x", b.MerkleRoot))
	lines = append(lines, fmt.Sprintf("    Timestamp:       %d", b.Timestamp))
	lines = append(lines, fmt.Sprintf("    Difficult:       %d", b.Difficult))
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

func (b *Block) GetCid() cid.Cid {
	//const DefaultHashFunction = uint64(mh.BLAKE2B_MIN + 31)
	nd, err := cbor.WrapObject(b, mh.SHA2_256, -1)
	if err != nil {
		panic(err)
	}

	return nd.Cid()
}

func (b *Block) GetHash() hash.Hash{
	return sha256.Sum256(b.Serialize())
}

func (b *Block) GetPrevHash() hash.Hash {
	return b.PrevHash
}

// SetHeight sets the height of the block
func (b *Block) SetHeight(height uint64) {
	b.Height = height
}
