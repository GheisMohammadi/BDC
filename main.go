package main

import (
	config "badcoin/src/config"
	logger "badcoin/src/helper/logger"
	node "badcoin/src/node"
	server "badcoin/src/server"
	storage "badcoin/src/storage"
	"context"
	"path/filepath"
)

func main() {
	// cli := new(cli.CLI)
	// cli.Run()

	//init logger
	logger.Init(false)
	logger.Info("logger initiated")

	//init configs
	logger.Info("loading configurations...")
	Configs, errConfig := config.Init("config")
	if errConfig != nil {
		panic("load configuration failed")
	}
	logger.Info("configurations loaded: ", Configs.Name)

	//init storage
	logger.Info("loading/installing storage...")
	storage := storage.InitStorage(Configs.Storage.Type)
	dataPath, _ := filepath.Abs("./data")
	errStorageInit := storage.Init(dataPath, Configs)
	if errStorageInit != nil {
		panic("storage failed")
	}
	//defer storage.Close()

	//create Node
	ctx := context.Background()
	node := node.CreateNewNode(ctx,Configs)

	//Start server
	srv := server.CreateNewServer(ctx,node,"3000")
	srv.ListenAndServe()

}
