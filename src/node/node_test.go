package node

import (
	"context"
	"testing"
)

func TestBlockchain(t *testing.T) {
	ctx := context.Background()
	testNode := CreateNewNode(ctx)
	testNode.StartMiner()
	t.Log(testNode)
}
