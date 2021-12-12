package blockchain

import (
	"context"
	"math"
	"math/big"
	"path/filepath"
	"time"

	config "badcoin/src/config"
	number "badcoin/src/helper/number"

	bitswap "github.com/ipfs/go-bitswap"
	network "github.com/ipfs/go-bitswap/network"
	blockservice "github.com/ipfs/go-blockservice"
	cid "github.com/ipfs/go-cid"
	dsleveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	nonerouting "github.com/ipfs/go-ipfs-routing/none"
	cbor "github.com/ipfs/go-ipld-cbor"
	multihash "github.com/multiformats/go-multihash"

	leveldb "github.com/syndtr/goleveldb/leveldb"

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
	BlockService blockservice.BlockService //block store to fetch blocks from nodes
	Blockstore   blockstore.Blockstore     //block store to fetch data locally
	BlockIndex   *leveldb.DB
	Accounts     *leveldb.DB
	Configs      *config.Configurations
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
	path := "../../data/" + configs.Storage.DBName + "_" + configs.ID + "_bs"
	fullpath, _ := filepath.Abs(path)
	//dstore := datastore.NewMapDatastore()
	dstore, err := dsleveldb.NewDatastore(fullpath, nil)
	if err != nil {
		panic(err)
	}

	// wrap the datastore in a 'content addressed blocks' layer
	chainblockstore := blockstore.NewBlockstore(dstore)

	//create block index db
	blockindexDBPath := "../../data/" + configs.Storage.DBName + "_" + configs.ID + "_bi"
	var errIndexDB error
	blockindex, errIndexDB := leveldb.OpenFile(blockindexDBPath, nil)
	if errIndexDB != nil {
		logger.Error(errIndexDB)
		panic(errIndexDB)
	}

	//Accounts db
	accPath := "../../data/" + configs.Storage.DBName + "_" + configs.ID + "_accs"
	accFullpath, _ := filepath.Abs(accPath)
	accDB, errAccDB := leveldb.OpenFile(accFullpath, nil)
	if errAccDB != nil {
		logger.Error(errAccDB)
		panic(errAccDB)
	}

	// now heres where it gets a bit weird. Its currently rather annoying to set up a bitswap instance.
	// Bitswap wants a datastore, and a 'network'. Bitswaps network instance
	// wants a libp2p node and a 'content routing' instance. We don't care
	// about content routing right now, so we want to give it a dummy one.
	// TODO: make bitswap easier to construct
	nr, _ := nonerouting.ConstructNilRouting(nil, nil, nil, nil)
	bsnet := network.NewFromIpfsHost(h, nr)

	bswap := bitswap.New(context.Background(), bsnet, chainblockstore)

	// Bitswap only fetches blocks from other nodes, to fetch blocks from
	// either the local cache, or a remote node, we can wrap it in a
	// 'blockservice'
	chainblockserviceice := blockservice.New(chainblockstore, bswap)

	genesis := CreateGenesisBlock(configs.Genesis.Nonce, configs.Genesis.Message)

	chain := &Blockchain{
		GenesisBlock: genesis,
		Head:         genesis,
		BlockService: chainblockserviceice,
		Blockstore:   chainblockstore,
		BlockIndex:   blockindex,
		Accounts:     accDB,
		Configs:      configs,
	}

	// make sure the genesis block is in our local blockstore
	chain.PutBlock(genesis)

	return chain
}

//LoadBlock loads block from local db or other nodes using block service
func (chain *Blockchain) LoadBlock(h *hash.Hash) (*block.Block, error) {

	bs := chain.BlockService
	bi := chain.BlockIndex

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

	//if block index is passed, store index in block index db
	if bi != nil {
		if err := chain.SaveBlockIndex(&out); err != nil {
			return nil, err
		}
	}

	return &out, nil
}

//PutBlock stores and broadcast block using block service and store it's index in block index db
func (chain *Blockchain) PutBlock(blk *block.Block) (*cid.Cid, error) {
	bs := chain.BlockService

	nd, err := cbor.WrapObject(blk, multihash.BLAKE2B_MIN+31, 32)
	if err != nil {
		return nil, err
	}

	if err := bs.AddBlock(context.Background(), nd); err != nil {
		return nil, err
	}

	err = chain.SaveBlockIndex(blk)
	if err != nil {
		return nil, err
	}

	cid := nd.Cid()
	return &cid, nil
}

func (chain *Blockchain) SaveBlockIndex(blk *block.Block) error {

	heightbytes := number.Int64ToByteArray(int64(blk.Height))
	hashbytes := blk.GetHashBytes()

	blkbytes, errGet := chain.BlockIndex.Get(hashbytes, nil)
	if errGet == nil {
		if len(blkbytes) > 0 {
			return nil
		}
	} else if errGet != nil {
		if errGet != leveldb.ErrNotFound {
			return errGet
		}
	}

	errUpdateAccounts := chain.UpdateAccounts(blk.Transactions)
	if errUpdateAccounts != nil {
		return errUpdateAccounts
	}

	err := chain.BlockIndex.Put(heightbytes, hashbytes, nil)
	if err != nil {
		return err
	}
	return nil
}

func CreateGenesisBlock(nonce int64, message string) *block.Block {
	//convert nonce to byte array
	//noncebytes := number.IntToHex(nonce)
	now := time.Now()
	tm := now.UnixMilli()

	genesisBlock := &block.Block{
		Height: 0,
		//Hash:	*hash.ZeroHash(),
		Header: block.BlockHeader{
			Version:    "0.0.1",
			PrevHash:   *hash.ZeroHash(),
			MerkleRoot: *hash.ZeroHash(),
			Timestamp:  tm,
			Nonce:      nonce,
			Miner:      "0x0",
			Difficulty: 1,
			Memo:       message,
		},
		TxsCount:     0,
		Reward:       new(big.Float).SetInt64(0),
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
		logger.Error("height (which is ", height, ") should be between 0 and ", chain.Head.Height, ".")
		return nil, errors.InvalidHeight
	}
	blockhashbytes, err := chain.BlockIndex.Get(number.Int64ToByteArray(int64(height)), nil) //chain.BlockIndex[height]
	if err != nil {
		if err == leveldb.ErrNotFound {
			logger.Error("block height ", height, " not found")
			return nil, nil
		}
		logger.Error("block height ", height, " fetch failed")
		return nil, err
	}
	blockhash, _ := hash.FromByteArray(blockhashbytes)
	if blockhash.String() == hash.ZeroHash().String() {
		logger.Error("block height ", height, " has zero hash")
		return nil, errors.InvalidHeight
	}
	return chain.LoadBlock(blockhash)
}

func validateTransactions(txs []*transaction.Transaction) bool {
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

func (chain *Blockchain) reload(oldBlock *block.Block, newBlock *block.Block) ([]*block.Block, error) {
	logger.Info("Rolling back...", newBlock)
	var newChain []*block.Block

	if oldBlock.GetHash().String() == newBlock.GetHash().String() {
		commonBlock := oldBlock
		logger.Info("Blockchain rolled back to block", commonBlock)
		return newChain, nil
	} else {
		newChain = append(newChain, newBlock)
		// Get the missing parent blocks by prevHash of newBlock
		prevBlock, err := chain.LoadBlock(&newBlock.Header.PrevHash)
		if err != nil {
			logger.Info("Fetching parent hashes of block failed -- aborting reload:", err)
			return nil, err
		}
		return chain.reload(oldBlock, prevBlock)
	}
}

//AddBlock adds block to blockchain
func (chain *Blockchain) AddBlock(blk *block.Block) *cid.Cid {
	if chain.ValidateBlock(blk) {
		prevhash := blk.Header.PrevHash.String()
		headhash := chain.Head.GetHash().String()
		if blk.Height > chain.Head.Height+1 && prevhash != headhash {
			// reload chain if prevhash is not chaintip hash
			chain.reload(chain.Head, blk)
		}
		blkCopy := *blk
		chain.Head = &blkCopy
		logger.Info("Block accepted, chain head set to block:", blkCopy.Hash) //string(blkCopy.Serialize()))
		cid, err := chain.PutBlock(&blkCopy)
		if err != nil {
			return nil
		}
		return cid
	}
	return nil
}

//SyncChain syncs chain from specific block (to genesis) using block service
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
		next, err := chain.LoadBlock(&fromhash)
		if err != nil {
			return err
		}

		cur = next
	}
}

func (bc *Blockchain) GetIterator() *Iterator {
	return &Iterator{
		bc.Head.Hash,
		bc,
	}
}

// GetBlockHashes returns a list of hashes with beginHash and maxNum limit
func (bc *Blockchain) GetBlockHashes(beginHash *hash.Hash, stopHash hash.Hash, maxNum int) ([]*hash.Hash, error) {
	var blocks []*hash.Hash
	bci := bc.GetIterator()
	err := bci.LocationHash(beginHash)
	if err != nil {
		return nil, err
	}

	getCount := 0
	for {
		block := bci.Next()
		h := block.GetHash()

		if stopHash.IsEqual(&h) {
			break
		}

		blocks = append(blocks, &h)
		getCount += 1

		if len(block.Header.PrevHash) == 0 {
			break
		}

		if getCount >= maxNum {
			break
		}
	}

	return blocks, nil
}

func (bc *Blockchain) AdjustDifficulty(blk *block.Block) int64 {
	blk.Header.Difficulty = 1
	return 1
}

func (bc *Blockchain) CalcReward(height uint64) *big.Float {

	reward := new(big.Float)

	hcat := int64(height / 100)
	halvingreward := 100 * math.Pow(0.5, float64(hcat))

	reward.SetFloat64(halvingreward)

	return reward
}
