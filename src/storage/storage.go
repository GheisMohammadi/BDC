package storage

import (
	config "badcoin/src/config"
	"badcoin/src/storage/level"
)

const (
	LEVEL_DB = 1 << iota
)

type Storage interface {
	Init(dataPath string, configs *config.Configurations) error
	Close() error
	GetDBFileName() string
	//RemoveBlock(blockHash []byte) error
	SaveBlockIndex(key, blockIndex []byte) error
	//SaveBlock(blockHash, blockData []byte) error
	//GetBlock(blockHash []byte) (blockData []byte, err error)
	//GetLastBlock() (lastHash, lastBlockData []byte, err error)
	//GetLashBlockHash() (lastHash []byte, err error)
	GetTXMemPool() ([]byte, error)
	SaveTXMemPool(txPool []byte) error
	CountTransactions() int
}

func InitStorage(storagetype uint8) Storage {
	switch storagetype {
	case LEVEL_DB:
		return new(level.Storage)
	default:
		panic("invalid storage type")
	}
}
