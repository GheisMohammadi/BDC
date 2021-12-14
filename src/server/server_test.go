package server

import (
	"context"
	"os"
	"testing"

	config "badcoin/src/config"
	node "badcoin/src/node"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	configs, _ := config.Init("")
	newNode := node.CreateNewNode(ctx, configs)
	server := CreateNewServer(ctx, newNode, "3000")
	if server==nil {
		t.Error("server creation failed")
	}
	server = nil
	//server.ListenAndServe()
	if err := os.RemoveAll("data"); err != nil {
		t.Error(err)
	}
}
