package blockchain

import (
	"context"
	"time"

	bitswap "github.com/ipfs/go-bitswap"
	network "github.com/ipfs/go-bitswap/network"
	blockservice "github.com/ipfs/go-blockservice"
	cid "github.com/ipfs/go-cid"
	datastore "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	nonerouting "github.com/ipfs/go-ipfs-routing/none"
	cbor "github.com/ipfs/go-ipld-cbor"
	multihash "github.com/multiformats/go-multihash"

	config "badcoin/src/config"
	number "badcoin/src/helper/number"

	block "badcoin/src/block"
	errors "badcoin/src/helper/error"
	hash "badcoin/src/helper/hash"
	logger "badcoin/src/helper/logger"
	transaction "badcoin/src/transaction"

	host "github.com/libp2p/go-libp2p-core/host"
)

type Blockchain struct {
	Head         *block.Block
	GenesisBlock *block.Block
	ChainDB      blockservice.BlockService //block store to fetch blocks from nodes
	Blockstore   blockstore.Blockstore     //block store to fetch data locally
	BlockIndex   map[uint64]hash.Hash
}

func Init() {
	// We need to Register our types with the cbor.
	// So, it pregenerates serializers for these types.
	cbor.RegisterCborType(block.Block{})
	cbor.RegisterCborType(transaction.Transaction{})
}

func NewBlockchain(h host.Host, configs *config.Configurations) *Blockchain {
	// base backing datastore, currently just in memory, but can be swapped out
	// easily for leveldb or other
	dstore := datastore.NewMapDatastore()

	// wrap the datastore in a 'content addressed blocks' layer
	blocks := blockstore.NewBlockstore(dstore)
	blockindex := make(map[uint64]hash.Hash)

	// now heres where it gets a bit weird. Its currently rather annoying to set up a bitswap instance.
	// Bitswap wants a datastore, and a 'network'. Bitswaps network instance
	// wants a libp2p node and a 'content routing' instance. We don't care
	// about content routing right now, so we want to give it a dummy one.
	// TODO: make bitswap easier to construct
	nr, _ := nonerouting.ConstructNilRouting(nil, nil, nil, nil)
	bsnet := network.NewFromIpfsHost(h, nr)

	bswap := bitswap.New(context.Background(), bsnet, blocks)

	// Bitswap only fetches blocks from other nodes, to fetch blocks from
	// either the local cache, or a remote node, we can wrap it in a
	// 'blockservice'
	blockserviceice := blockservice.New(blocks, bswap)

	genesis := CreateGenesisBlock(configs.Genesis.Nonce)

	// make sure the genesis block is in our local blockstore
	PutBlock(blockserviceice, blockindex, genesis)

	return &Blockchain{
		GenesisBlock: genesis,
		Head:         genesis,
		ChainDB:      blockserviceice,
		Blockstore:   blocks,
		BlockIndex:   blockindex,
	}
}

func LoadBlock(bs blockservice.BlockService, h *hash.Hash) (*block.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	blkcid, _ := h.ToCid()
	data, err := bs.GetBlock(ctx, *blkcid)
	if err != nil {
		return nil, err
	}

	var out block.Block
	if err := cbor.DecodeInto(data.RawData(), &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func PutBlock(bs blockservice.BlockService, bi map[uint64]hash.Hash, blk *block.Block) (*cid.Cid, error) {
	nd, err := cbor.WrapObject(blk, multihash.BLAKE2B_MIN+31, 32)
	if err != nil {
		return nil, err
	}

	if err := bs.AddBlock(context.Background(), nd); err != nil {
		return nil, err
	}
	bi[blk.Height] = blk.GetHash()
	cid := nd.Cid()
	return &cid, nil
}

func CreateGenesisBlock(nonce int64) *block.Block {
	//convert nonce to byte array
	noncebytes := number.IntToHex(nonce)
	now := time.Now()
	tm := now.UnixMilli()

	genesisBlock := &block.Block{
		Height: 0,
		//Hash:	*hash.ZeroHash(),
		Header: block.BlockHeader{
			Version:    10000,
			PrevHash:   *hash.ZeroHash(),
			MerkleRoot: *hash.ZeroHash(),
			Timestamp:  uint64(tm),
			Nonce:      noncebytes,
			Difficulty: 1,
			Solution:   "",
		},
		TxsCount:     0,
		Transactions: nil,
	}
	genesisBlock.UpdateHash()

	return genesisBlock
}

func (chain *Blockchain) GetChainTip() *block.Block {
	return chain.Head
}

func (chain *Blockchain) GetBlock(height uint64) (*block.Block, error) {
	if height == 0 {
		return chain.GenesisBlock, nil
	}
	if height < 0 || height > chain.Head.Height {
		return nil, errors.InvalidHeight
	}
	blockhash := chain.BlockIndex[height]
	if blockhash.String() == hash.ZeroHash().String() {
		return nil, errors.InvalidHeight
	}
	return LoadBlock(chain.ChainDB, &blockhash)
}

func validateTransactions(txs []transaction.Transaction) bool {
	// TODO:Validate tx format and logic
	return true
}

// 1- Check that prevHash of new block (it should be equal to hash of chainTip)
// 2- Validate Transactions
// 3- Time is greater than time of chainTip
func (chain *Blockchain) ValidateBlock(blk *block.Block) bool {
	chainTip := chain.Head
	if blk.Height <= chainTip.Height {
		logger.Info("Block validation failed: Height is less than chaintip")
		return false
	}
	tipHash := chainTip.GetHash()
	if !blk.Header.PrevHash.IsEqual(&tipHash) {
		logger.Info("Block validation failed: Invalid PrevHash")
		return false
	}
	if !validateTransactions(blk.Transactions) {
		logger.Info("Block validation failed: Block Contains invalid tx")
		return false
	}
	if blk.Header.Timestamp < chainTip.Header.Timestamp {
		logger.Info("Block validation failed: Invalid Time")
		return false
	}
	return true
}

func (chain *Blockchain) rollback(oldBlock *block.Block, newBlock *block.Block) ([]*block.Block, error) {
	logger.Info("Rolling back...", newBlock)
	var newChain []*block.Block

	if oldBlock.GetHash().String() == newBlock.GetHash().String() {
		commonBlock := oldBlock
		logger.Info("Blockchain rolled back to block", commonBlock)
		return newChain, nil
	} else {
		newChain = append(newChain, newBlock)
		// Get the missing parent blocks by prevHash of newBlock
		prevBlock, err := LoadBlock(chain.ChainDB, &newBlock.Header.PrevHash)
		if err != nil {
			logger.Info("Fetching parent hashes of block failed -- aborting rollback:", err)
			return nil, err
		}
		return chain.rollback(newBlock, prevBlock)
	}
}

func (chain *Blockchain) AddBlock(blk *block.Block) *cid.Cid {
	if chain.ValidateBlock(blk) {
		prevhash := blk.Header.PrevHash.String()
		headhash := chain.Head.GetHash().String()
		if blk.Height > chain.Head.Height+1 && prevhash != headhash {
			// rollback chain if prevhash is not chaintip hash
			chain.rollback(chain.Head, blk)
		}
		blkCopy := *blk
		chain.Head = &blkCopy
		logger.Info("Block accepted, chain head set to block:", string(blkCopy.Serialize()))
		cid, err := PutBlock(chain.ChainDB, chain.BlockIndex, &blkCopy)
		if err != nil {
			return nil
		}
		return cid
	}
	return nil
}

func (chain *Blockchain) SyncChain(from *block.Block) error {
	cur := from
	for {
		prevcid, _ := cur.Header.PrevHash.ToCid()
		haveParent, err := chain.Blockstore.Has(context.Background(), *prevcid)
		if err != nil {
			return err
		}

		if haveParent {
			return nil
		}

		fromhash := from.Header.PrevHash
		next, err := LoadBlock(chain.ChainDB, &fromhash)
		if err != nil {
			return err
		}

		cur = next
	}
}
