package level

import (
	config "badcoin/src/config"
	errors "badcoin/src/helper/error"
	logger "badcoin/src/helper/logger"
	"log"

	leveldb "github.com/syndtr/goleveldb/leveldb"
)

const (
	LastHashKey  = "LastBlockHashKey"
	TxMemPoolKey = "TxMemPool"
)

type Storage struct {
	DBName string

	BlocksDB     *leveldb.DB
	UTXODB       *leveldb.DB
	TXsMemPoolDB *leveldb.DB
	BlockIndexDB *leveldb.DB
	StatsDB      *leveldb.DB
}

func (lvldb *Storage) Init(dataPath string, configs *config.Configurations) error {
	//initiate data path and db name
	lvldb.DBName = configs.Storage.DBName + "_" + configs.ID
	logger.Info("db name: ",lvldb.DBName)
	//create/open blocks db
	blocksDBName := lvldb.DBName + "_" + configs.Storage.Collections.Blocks
	var errBlocksDB error
	lvldb.BlocksDB, errBlocksDB = leveldb.OpenFile(dataPath+"/"+blocksDBName, nil)
	if errBlocksDB != nil {
		logger.Error(errBlocksDB)
		return errors.StorageInitFailed
	}

	//create/open utxo db
	utxoDBName := lvldb.DBName + "_" + configs.Storage.Collections.UTXO
	var errUTXOsDB error
	lvldb.UTXODB, errUTXOsDB = leveldb.OpenFile(dataPath+"/"+utxoDBName, nil)
	if errUTXOsDB != nil {
		logger.Error(errUTXOsDB)
		return errors.StorageInitFailed
	}

	//create/open mempool db
	mempoolDBName := lvldb.DBName + "_" + configs.Storage.Collections.TXsMemPool
	var errMempoolDB error
	lvldb.TXsMemPoolDB, errMempoolDB = leveldb.OpenFile(dataPath+"/"+mempoolDBName, nil)
	if errMempoolDB != nil {
		logger.Error(errMempoolDB)
		return errors.StorageInitFailed
	}

	//create/open blocks index db
	blocksindexDBName := lvldb.DBName + "_" + configs.Storage.Collections.BlockIndex
	var errBlocksIndexDB error
	lvldb.BlockIndexDB, errBlocksIndexDB = leveldb.OpenFile(dataPath+"/"+blocksindexDBName, nil)
	if errBlocksIndexDB != nil {
		logger.Error(errBlocksIndexDB)
		return errors.StorageInitFailed
	}

	//create/open stats db
	statsDBName := lvldb.DBName + "_" + configs.Storage.Collections.Stats
	var errStatsDB error
	lvldb.StatsDB, errStatsDB = leveldb.OpenFile(dataPath+"/"+statsDBName, nil)
	if errStatsDB != nil {
		logger.Error(errStatsDB)
		return errors.StorageInitFailed
	}

	return nil
}

func (lvldb *Storage) Close() error {
	// lvldb.BlocksDB.Close()
	// lvldb.UTXODB.Close()
	// lvldb.TXsMemPoolDB.Close()
	// lvldb.BlockIndexDB.Close()
	// lvldb.StatsDB.Close()

	return nil
}

//GetDBFileName returns db name
func (lvldb *Storage) GetDBFileName() string {
	return lvldb.DBName
}

// RemoveBlock remove block
func (lvldb *Storage) RemoveBlock(blockHash []byte) error {
	err := lvldb.BlocksDB.Delete(blockHash, nil)
	return err
}

// SaveBlockIndex save block index to db
func (lvldb *Storage) SaveBlockIndex(key, blockIndex []byte) error {
	err := lvldb.BlockIndexDB.Put(key, blockIndex, nil)
	return err
}

func (lvldb *Storage) SaveBlock(blockHash, blockData []byte) error {
	err := lvldb.BlocksDB.Put(blockHash, blockData, nil)
	if err != nil {
		return err
	}
	err = lvldb.StatsDB.Put([]byte(LastHashKey), blockHash, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetBlock query block data with block hash
// if not exists, return ErrorBlockNotFount
func (lvldb *Storage) GetBlock(blockHash []byte) (blockData []byte, err error) {

	blockData, err = lvldb.BlocksDB.Get(blockHash, nil)
	if blockData == nil || err != nil {
		return nil, errors.BlockNotFount
	}
	return blockData, nil

}

func (lvldb *Storage) GetLastBlock() (lastHash, lastBlockData []byte, err error) {
	lastBlockHash, errLastBlock := lvldb.StatsDB.Get([]byte(LastHashKey), nil)
	if errLastBlock != nil {
		return nil, nil, errLastBlock
	}
	lastBlockFullData, errBlockData := lvldb.BlocksDB.Get(lastHash, nil)
	if errBlockData != nil {
		return nil, nil, errBlockData
	}
	return lastBlockHash, lastBlockFullData, err
}

func (lvldb *Storage) GetLashBlockHash() (lastHash []byte, err error) {
	lastHash, _, err = lvldb.GetLastBlock()
	return
}

func (lvldb *Storage) GetTXMemPool() ([]byte, error) {
	txPool, err := lvldb.TXsMemPoolDB.Get([]byte(TxMemPoolKey), nil)
	return txPool, err
}

// SaveTXMemPool save mempool into db
func (lvldb *Storage) SaveTXMemPool(txPool []byte) error {
	errSaveMempool := lvldb.TXsMemPoolDB.Put([]byte(TxMemPoolKey), txPool, nil)
	return errSaveMempool
}

func (lvldb *Storage) CountTransactions() int {
	counter := 0

	iter := lvldb.UTXODB.NewIterator(nil, nil)
	for iter.Next() {
		//key := iter.Key()
		//value := iter.Value()
		counter++
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Panic(err)
	}

	return counter
}
