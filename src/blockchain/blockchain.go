package blockchain

import (
	"context"
	"math"
	"math/big"
	"path/filepath"
	"time"

	config "badcoin/src/config"
	number "badcoin/src/helper/number"

	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	//graphnet "github.com/ipfs/go-graphsync/network"
	blockservice "github.com/ipfs/go-blockservice"
	cid "github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"

	//nonerouting "github.com/ipfs/go-ipfs-routing/none"
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
	cbor.RegisterCborType(big.Float{})
	cbor.RegisterCborType(block.BlockHeader{})
	cbor.RegisterCborType(block.Block{})
	cbor.RegisterCborType(transaction.Transaction{})
}

//Find latest Height
//Find Head
func LoadBlockchain(chainblockstore blockstore.Blockstore) (*block.Block, *block.Block, error) {
	ctx := context.TODO()
	keychan, _ := chainblockstore.AllKeysChan(ctx)
	lastHeight := uint64(0)
	var head *block.Block
	var genesis *block.Block
	loadedblocks := 0
	for {
		select {
		case k, ok := <-keychan:
			if ok == false {
				logger.Info("iterating block store completed! ",loadedblocks," blocks are loaded")
				return head, genesis, nil
			}
			if k.ByteLen() == 0 {
				break
			} else {
				blkcid, _ := cid.Decode(k.String())
				data, err := chainblockstore.Get(ctx, blkcid)
				if err != nil {
					return nil, nil, err
				}
				var blk block.Block
				if err := cbor.DecodeInto(data.RawData(), &blk); err != nil {
					return nil, nil, err
				}
				loadedblocks++
				if blk.Height > lastHeight {
					lastHeight = blk.Height
					head = &blk
				}
				if blk.Height == 0 {
					genesis = &blk
				}
			}
		case <-ctx.Done():
			logger.Info("loading is done!")
			return head, genesis, nil
		}
	}

}

func NewBlockchain(h host.Host, chainblockstore blockstore.Blockstore, bswap exchange.Interface, configs *config.Configurations) *Blockchain {
	//create block index db
	blockindexDBPath := "data/" + configs.Storage.DBName + "_" + configs.ID + "_bi"
	var errIndexDB error
	blockindex, errIndexDB := leveldb.OpenFile(blockindexDBPath, nil)
	if errIndexDB != nil {
		logger.Error(errIndexDB)
		panic(errIndexDB)
	}

	//Accounts db
	accPath := "data/" + configs.Storage.DBName + "_" + configs.ID + "_accs"
	accFullpath, _ := filepath.Abs(accPath)
	accDB, errAccDB := leveldb.OpenFile(accFullpath, nil)
	if errAccDB != nil {
		logger.Error(errAccDB)
		panic(errAccDB)
	}

	// Bitswap only fetches blocks from other nodes, to fetch blocks from
	// either the local cache, or a remote node, we can wrap it in a
	// 'blockservice'
	chainblockserviceice := blockservice.NewWriteThrough(chainblockstore, bswap)

	//load blockchain
	curhead, curgenesis, errLoad := LoadBlockchain(chainblockstore)
	if errLoad != nil {
		logger.Error(errLoad)
		panic(errLoad)
	}

	var genesis *block.Block
	var head *block.Block
	if curhead == nil {
		logger.Info("creating genesis block ...")
		genesis = CreateGenesisBlock(configs.Genesis.Nonce, configs.Genesis.Message)
		head = genesis
	} else {
		logger.Info("recovered current stored chain. Head is on the height: ", curhead.Height)
		genesis = curgenesis
		head = curhead
	}

	chain := &Blockchain{
		GenesisBlock: genesis,
		Head:         head,
		BlockService: chainblockserviceice,
		Blockstore:   chainblockstore,
		BlockIndex:   blockindex,
		Accounts:     accDB,
		Configs:      configs,
	}

	// make sure the genesis block is in our local blockstore
	chain.PutBlock(genesis)

	isonline := bswap.IsOnline()
	logger.Info("exchange online is ", isonline)
	return chain
}

//LoadBlock loads block from local db or other nodes using block service
func (chain *Blockchain) LoadBlock(blkcid *cid.Cid) (*block.Block, error) {

	if blkcid == nil {
		return nil, nil
	}

	bsrv := chain.BlockService
	bi := chain.BlockIndex

	ctx, cancel := context.WithCancel(context.Background()) //, time.Second*10)
	defer cancel()

	ok, _ := bsrv.Blockstore().Has(ctx, *blkcid)
	if ok == false {
		logger.Error("Block is not exist in block store, trying to fetch from exchange...")
		isonline := bsrv.Exchange().IsOnline()
		if isonline == false {
			logger.Error("Exchange is not online to retrieve block")
			return nil, errors.ExchangeISNotOnline
		}
	}
	//sess := blockservice.NewSession(ctx, bsrv)

	data, err := bsrv.GetBlock(ctx, *blkcid)
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
	bsrv := chain.BlockService

	nd, err := cbor.WrapObject(blk, multihash.BLAKE2B_MIN+31, 32)
	if err != nil {
		return nil, err
	}

	if err := bsrv.AddBlock(context.Background(), nd); err != nil {
		return nil, err
	}

	err = bsrv.Exchange().HasBlock(context.Background(), nd)
	if err != nil {
		return nil, err
	}

	err = chain.SaveBlockIndex(blk)
	if err != nil {
		return nil, err
	}

	cid := nd.Cid()
	return &cid, nil
}

// Height ---- Map to ----> Block Cid
func (chain *Blockchain) SaveBlockIndex(blk *block.Block) error {

	heightbytes := number.Int64ToByteArray(int64(blk.Height))
	hashbytes := chain.GetBlockCid(blk).Bytes() //blk.GetHashBytes()

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
		PrevCid:      nil,
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
	blkCidbytes, err := chain.BlockIndex.Get(number.Int64ToByteArray(int64(height)), nil) //chain.BlockIndex[height]
	if err != nil {
		if err == leveldb.ErrNotFound {
			logger.Error("block height ", height, " not found")
			return nil, nil
		}
		logger.Error("block height ", height, " fetch failed")
		return nil, err
	}
	// blockhash, _ := hash.FromByteArray(blockhashbytes)
	// if blockhash.String() == hash.ZeroHash().String() {
	// 	logger.Error("block height ", height, " has zero hash")
	// 	return nil, errors.InvalidHeight
	// }
	_, blkcid, _ := cid.CidFromBytes(blkCidbytes)
	return chain.LoadBlock(&blkcid)
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
	// tipHash := chainTip.GetHash()
	// if !blk.Header.PrevHash.IsEqual(&tipHash) {
	// 	logger.Info("Block validation failed: Invalid PrevHash")
	// 	return false
	// }
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
	logger.Info("Reloading chain from block: ", newBlock.Height)
	var newChain []*block.Block

	if oldBlock.GetHash().String() == newBlock.GetHash().String() {
		commonBlock := oldBlock
		logger.Info("Blockchain reloaded back from block", commonBlock)
		return newChain, nil
	} else {
		newChain = append(newChain, newBlock)
		logger.Info("new block added to chain: ", string(newBlock.Serialize()))
		// Get the missing parent blocks by prevHash of newBlock
		if newBlock.Height == 0 {
			return newChain, nil
		}
		logger.Info("fetching block height ", newBlock.Height-1, " for cid: ", newBlock.PrevCid.String())
		prevBlock, err := chain.LoadBlock(newBlock.PrevCid)
		if err != nil {
			logger.Error("Fetching parent hashes of block failed -- aborting reload:", err)
			return nil, err
		}
		logger.Info("block height ", newBlock.Height-1, " loaded!")
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
			logger.Info("Reloading blocks from height: ", blk.Height)
			_, errReload := chain.reload(chain.Head, blk)
			if errReload != nil {
				return nil
			}
		}
		blkCopy := *blk
		chain.Head = &blkCopy
		logger.Info("Block accepted, chain head set to block height:", blkCopy.Height) //string(blkCopy.Serialize()))
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
		prevcid := cur.PrevCid
		haveParent, err := chain.Blockstore.Has(context.Background(), *prevcid)
		if err != nil {
			return err
		}

		if haveParent {
			return nil
		}

		fromhash := from.PrevCid
		next, err := chain.LoadBlock(fromhash)
		if err != nil {
			return err
		}

		cur = next
	}
}

func (bc *Blockchain) GetIterator() *Iterator {
	return &Iterator{
		bc.GetBlockCid(bc.Head),
		bc,
	}
}

// GetBlockHashes returns a list of hashes with beginHash and maxNum limit
func (bc *Blockchain) GetBlockCids(beginCid *cid.Cid, stopCid cid.Cid, maxNum int) ([]*cid.Cid, error) {
	var blocks []*cid.Cid
	bci := bc.GetIterator()
	err := bci.LocationHash(beginCid)
	if err != nil {
		return nil, err
	}

	getCount := 0
	for {
		block := bci.Next()
		h := bc.GetBlockCid(block)

		if stopCid.Equals(*h) {
			break
		}

		blocks = append(blocks, h)
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

func (bc *Blockchain) GetBlockCid(b *block.Block) *cid.Cid {
	nd, err := cbor.WrapObject(*b, multihash.BLAKE2B_MIN+31, 32)
	if err != nil {
		panic(err)
	}
	bcid := nd.Cid()
	return &bcid
}
