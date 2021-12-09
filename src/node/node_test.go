package node

import (
	config "badcoin/src/config"
	"context"
	"testing"
)

func TestBlockchain(t *testing.T) {
	ctx := context.Background()
	configs, _ := config.Init("")
	testNode := CreateNewNode(ctx, configs)
	testNode.StartMiner()
	t.Log(testNode)
}
