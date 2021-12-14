package node

import (
	config "badcoin/src/config"
	"context"
	"os"
	"testing"
)

func TestBlockchain(t *testing.T) {
	ctx := context.Background()
	configs, _ := config.Init("")
	testNode := CreateNewNode(ctx, configs)
	testNode.StartMiner()
	t.Log(testNode)

	if err := os.RemoveAll("data"); err != nil {
		t.Error(err)
	}
}
