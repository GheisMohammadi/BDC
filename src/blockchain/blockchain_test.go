package blockchain

import (
	"context"
	"os"
	"testing"

	config "badcoin/src/config"

	bitswap "github.com/ipfs/go-bitswap"
	network "github.com/ipfs/go-bitswap/network"
	datastore "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	nonerouting "github.com/ipfs/go-ipfs-routing/none"
	libp2p "github.com/libp2p/go-libp2p"
)

func TestBlockchain(t *testing.T) {
	configs, _ := config.Init("")

	h, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		panic(err)
	}

	bs := blockstore.NewBlockstore(datastore.NewMapDatastore())
	nr, _ := nonerouting.ConstructNilRouting(context.Background(), nil, nil, nil)
	net := network.NewFromIpfsHost(h, nr)
	bswap := bitswap.New(context.Background(), net, bs)

	bc := NewBlockchain(h, bs, bswap, configs)

	t.Log(bc.Head.Height)

	if err := os.RemoveAll("data"); err != nil {
		t.Error(err)
	}
}
