package blockchain

import (
	"testing"

	libp2p "github.com/libp2p/go-libp2p"
	config "badcoin/src/config"
)

func TestBlockchain(t *testing.T) {
	
	configs,_ := config.Init("")
	h, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		panic(err)
	}
	bc := NewBlockchain(h,configs)
	t.Log(bc.Head.Height)
}
