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

	block "badcoin/src/block"
	logger "badcoin/src/helper/logger"
	transaction "badcoin/src/transaction"

	host "github.com/libp2p/go-libp2p-core/host"
)

type Blockchain struct {
	Head         *block.Block
	GenesisBlock *block.Block
	ChainDB      blockservice.BlockService //block store to fetch blocks from nodes
	Blockstore   blockstore.Blockstore     //block store to fetch data locally
}

func Init() {
	// We need to Register our types with the cbor.
	// So, it pregenerates serializers for these types.
	cbor.RegisterCborType(block.Block{})
	cbor.RegisterCborType(transaction.Transaction{})
}

func NewBlockchain(h host.Host) *Blockchain {
	// base backing datastore, currently just in memory, but can be swapped out
	// easily for leveldb or other
	dstore := datastore.NewMapDatastore()

	// wrap the datastore in a 'content addressed blocks' layer
	blocks := blockstore.NewBlockstore(dstore)

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

	genesis := CreateGenesisBlock()

	// make sure the genesis block is in our local blockstore
	PutBlock(blockserviceice, genesis)

	return &Blockchain{
		GenesisBlock: genesis,
		Head:         genesis,
		ChainDB:      blockserviceice,
		Blockstore:   blocks,
	}
}

func LoadBlock(bs blockservice.BlockService, c *cid.Cid) (*block.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	data, err := bs.GetBlock(ctx, *c)
	if err != nil {
		return nil, err
	}

	var out block.Block
	if err := cbor.DecodeInto(data.RawData(), &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func PutBlock(bs blockservice.BlockService, blk *block.Block) (*cid.Cid, error) {
	nd, err := cbor.WrapObject(blk, multihash.BLAKE2B_MIN+31, 32)
	if err != nil {
		return nil, err
	}

	if err := bs.AddBlock(context.Background(), nd); err != nil {
		return nil, err
	}
	cid := nd.Cid()
	return &cid, nil
}

func CreateGenesisBlock() *block.Block {
	genesisBlock := &block.Block{
		Height: 0,
		Time:   42,
	}
	return genesisBlock
}

func (chain *Blockchain) GetChainTip() *block.Block {
	return chain.Head
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
	if !blk.PrevHash.Equals(chainTip.GetCid()) {
		logger.Info("Block validation failed: Invalid PrevHash")
		return false
	}
	if !validateTransactions(blk.Transactions) {
		logger.Info("Block validation failed: Block Contains invalid tx")
		return false
	}
	if blk.Time < chainTip.Time {
		logger.Info("Block validation failed: Invalid Time")
		return false
	}
	return true
}

func (chain *Blockchain) rollback(oldBlock *block.Block, newBlock *block.Block) ([]*block.Block, error) {
	logger.Info("Rolling back...", newBlock)
	var newChain []*block.Block

	if oldBlock.GetCid() == newBlock.GetCid() {
		commonBlock := oldBlock
		logger.Info("Blockchain rolled back to block", commonBlock)
		return newChain, nil
	} else {
		newChain = append(newChain, newBlock)
		// Get the missing parent blocks by prevHash of newBlock
		prevBlock, err := LoadBlock(chain.ChainDB, newBlock.PrevHash)
		if err != nil {
			logger.Info("Fetching parent hashes of block failed -- aborting rollback:", err)
			return nil, err
		}
		return chain.rollback(newBlock, prevBlock)
	}
}

func (chain *Blockchain) AddBlock(blk *block.Block) *cid.Cid {
	if chain.ValidateBlock(blk) {
		if blk.Height > chain.Head.Height+1 && *blk.PrevHash != chain.Head.GetCid() {
			// rollback chain if prevhash is not chaintip hash
			chain.rollback(chain.Head, blk)
		}
		blkCopy := *blk
		chain.Head = &blkCopy
		logger.Info("Block accepted, chain head set to block:", string(blkCopy.Serialize()))
		cid, err := PutBlock(chain.ChainDB, &blkCopy)
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
		haveParent, err := chain.Blockstore.Has(context.Background(), *cur.PrevHash)
		if err != nil {
			return err
		}

		if haveParent {
			return nil
		}

		next, err := LoadBlock(chain.ChainDB, from.PrevHash)
		if err != nil {
			return err
		}

		cur = next
	}
}
