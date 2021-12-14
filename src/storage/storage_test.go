package storage

import (
	//level "badcoin/src/storage/level"
	config "badcoin/src/config"
	"badcoin/src/storage/level"
	"path/filepath"
	"testing"
	//"io/ioutil"
)

func Test_Leveldb_Storage(t *testing.T) {

	//init configs
	configs,errConfigs := config.Init("")
	t.Log(configs)
	if errConfigs!=nil {
		t.Error(errConfigs)
	}
	t.Log(configs.Storage.DBName)

	dataPath, _ := filepath.Abs("../../data")

	lvlStorage := new(level.Storage)
	errInit := lvlStorage.Init(dataPath,configs)
	if errInit != nil {
		t.Error(errInit)
	}

	//lvlStorage.SaveBlock([]byte("block hash"),[]byte("block data"))

	// blockdata,errBlock := lvlStorage.GetBlock([]byte("block hash"))
	// if errBlock!=nil {
	// 	t.Error(errBlock)
	// }
	// t.Log(string(blockdata[:]))

	lvlStorage.Close()
}
