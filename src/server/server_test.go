package server

import (
	"context"
	"testing"

	node "badcoin/src/node"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	newNode := node.CreateNewNode(ctx)
	server := CreateNewServer(ctx, newNode, "3000")
	server.ListenAndServe()
}
