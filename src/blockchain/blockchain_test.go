package blockchain

import (
	"testing"

	libp2p "github.com/libp2p/go-libp2p"
)

func TestBlockchain(t *testing.T) {
	h, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		panic(err)
	}
	bc := NewBlockchain(h)
	t.Log(bc.Head.Height)
}
