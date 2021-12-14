package block

import (
	hash "badcoin/src/helper/hash"
	"badcoin/src/transaction"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/go-cid"
)

type BlockMessage struct {
	Height  []int64 `json:"height"`
	Message string  `json:"message"`
}

type BlockMessages struct {
	Messages []BlockMessage `json:"Messages"`
}

type BlockHeader struct {
	Version    string
	PrevHash   hash.Hash
	MerkleRoot hash.Hash
	Timestamp  int64
	Nonce      int64
	Miner      string
	Difficulty uint32
	Memo       string
}

type Block struct {
	Height       uint64
	Hash         hash.Hash
	PrevCid      *cid.Cid
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

func ReadBlockMessage(height uint64, messags_json_path string) (string, error) {

	messagefile, _ := filepath.Abs("config/block_message.json")
	if len(messags_json_path) > 0 {
		messagefile = messags_json_path
	}
	//logger.Info(messagefile)
	jsonFile, err := os.Open(messagefile)
	if err != nil {
		return "", err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var memos BlockMessages

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &memos)
	if err != nil {
		return "", err
	}

	// we iterate through every user within our users array and
	// print out the user Type, their name, and their facebook url
	// as just an example
	for i := 0; i < len(memos.Messages); i++ {
		heights := memos.Messages[i].Height
		message := memos.Messages[i].Message

		if len(heights) == 1 {
			if heights[0] == int64(height) {
				return message, nil
			}
		} else if len(heights) == 2 {
			if int64(height) >= heights[0] && (int64(height) <= heights[1] || heights[1] == -1) {
				return message, nil
			}
		} else if len(heights) > 2 {
			for _, h := range heights {
				if height == uint64(h) {
					return message, nil
				}
			}

		}
	}

	return "", nil
}
