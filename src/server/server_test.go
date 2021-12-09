package server

import (
	"context"
	"testing"

	config "badcoin/src/config"
	node "badcoin/src/node"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	configs, _ := config.Init("")
	newNode := node.CreateNewNode(ctx, configs)
	server := CreateNewServer(ctx, newNode, "3000")
	server.ListenAndServe()
}
