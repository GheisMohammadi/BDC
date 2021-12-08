package block

import (
	"badcoin/src/transaction"
	"crypto/sha256"
	"encoding/json"

	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	mh "github.com/multiformats/go-multihash"
)

type Block struct {
	PrevHash     *cid.Cid
	Transactions []transaction.Transaction
	Height       uint64
	Time         uint64
	Nonce        []byte
	Solution     string
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

func (b *Block) GetHash() [32]byte {
	return sha256.Sum256(b.Serialize())
}
