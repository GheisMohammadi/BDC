package node

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	block "badcoin/src/block"
	blockchain "badcoin/src/blockchain"
	logger "badcoin/src/helper/logger"
	mempool "badcoin/src/mempool"
	transaction "badcoin/src/transaction"
	wallet "badcoin/src/wallet"

	config "badcoin/src/config"

	proofofwork "badcoin/src/pow"

	"github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	floodsub "github.com/libp2p/go-libp2p-pubsub"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	bitswap "github.com/ipfs/go-bitswap"
	network "github.com/ipfs/go-bitswap/network"
	"github.com/ipfs/go-datastore"

	//graphnet "github.com/ipfs/go-graphsync/network"

	blockstore "github.com/ipfs/go-ipfs-blockstore"
	ldbopts "github.com/syndtr/goleveldb/leveldb/opt"

	//nonerouting "github.com/ipfs/go-ipfs-routing/none"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	routing "github.com/libp2p/go-libp2p-core/routing"

	dsleveldb "github.com/ipfs/go-ds-leveldb"
)

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "badcoin-network"

type Node struct {
	p2pNode    host.Host
	mempool    *mempool.Mempool
	blockchain *blockchain.Blockchain
	pubsub     *floodsub.PubSub
	wallet     *wallet.Wallet
	pow        *proofofwork.ProofOfWork
}

func DHTRoutingFactory() func(host.Host) (routing.PeerRouting, error) {
	makeRouting := func(h host.Host) (routing.PeerRouting, error) {
		dhtInst, err := dht.New(context.Background(), h) //, opts...)
		if err != nil {
			return nil, err
		}
		//d.dht = dhtInst
		return dhtInst, nil
	}

	return makeRouting
}

//var router routing.Routing
func makeDHT(h host.Host) (routing.Routing, error) {
	// mode := dht.ModeServer
	// opts := []dht.Option{dht.Mode(mode),
	// 	// 	//dht.Datastore(repo.ChainDatastore()),
	// 	// 	//dht.NamespacedValidator("v", validator),
	// 	//dht.ProtocolPrefix(net.FilecoinDHT(networkName)),
	// 	// 	dht.QueryFilter(dht.PublicQueryFilter),
	// 	dht.Datastore(datastore.NewMapDatastore()),
	// 	dht.RoutingTableFilter(dht.PublicRoutingTableFilter),
	// 	dht.DisableProviders(),
	// 	dht.DisableValues(),
	// }
	// r, err := dht.New(
	// 	context.Background(), h, opts...,
	// )

	// if err != nil {
	// 	return nil, err//errors.Wrap(err, "failed to setup routing")
	// }

	//r, err := dht.New(context.Background(), h, opts...)
	r := dht.NewDHT(context.Background(), h, datastore.NewMapDatastore())

	//r := dht.NewDHT(context.Background(),h,datastore.NewMapDatastore())
	// if err != nil {
	// 	logger.Error(err)
	// 	return nil, err
	// }
	return r, nil
}

func CreateNewNode(ctx context.Context, configs *config.Configurations) *Node {
	var node Node

	var opts []libp2p.Option
	opts = append(opts, libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	opts = append(opts, libp2p.Routing(DHTRoutingFactory()))
	//router, _ := makeDHT(h)
	//nr, _ := nonerouting.ConstructNilRouting(context.TODO(), nil, nil, nil)

	newNode, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	pubsub, err := floodsub.NewFloodSub(ctx, newNode)
	if err != nil {
		panic(err)
	}

	// base backing datastore, currently just in memory, but can be swapped out
	// easily for leveldb or other
	path := "../../data/" + configs.Storage.DBName + "_" + configs.ID + "_bs"
	fullpath, _ := filepath.Abs(path)
	//dstore := datastore.NewMapDatastore()
	dstore, err := dsleveldb.NewDatastore(fullpath, &dsleveldb.Options{
		Compression: ldbopts.NoCompression,
		NoSync:      false,
		Strict:      ldbopts.StrictAll,
	})
	if err != nil {
		panic(err)
	}

	// wrap the datastore in a 'content addressed blocks' layer
	chainblockstore := blockstore.NewBlockstore(dstore)


	ccc,_:=cid.Decode("bafk2bzaced7scnipal3tpzll2xxuxcwzxvj53agdfphh57ifpk73pn5hhg3n2")
	has,_:=chainblockstore.Has(context.Background(),ccc)
	logger.Info("HAS:",has)

	router, _ := makeDHT(newNode)
	//nr, _ := nonerouting.ConstructNilRouting(context.TODO(), nil, nil, nil)

	//var router routing.ContentRouting
	//rot := gnet.NewFromLibp2pHost(h)
	net := network.NewFromIpfsHost(newNode, router) //, network.Prefix("/bdcchain"))

	//bitswapOptions := []bitswap.Option{bitswap.ProvideEnabled(true)}
	bswap := bitswap.New(context.Background(), net, chainblockstore) //, bitswapOptions...)

	// setup local mDNS discovery
	if err := setupDiscovery(newNode); err != nil {
		panic(err)
	}

	for i, addr := range newNode.Addrs() {
		logger.Info(i, ": ", addr.String()+"/ipfs/"+newNode.ID().Pretty())
	}

	if len(os.Args) > 1 {
		addrstr := os.Args[1]
		addr, err := ipfsaddr.ParseString(addrstr)
		if err == nil {
			pInfo, _ := peer.AddrInfoFromP2pAddr(addr.Multiaddr())

			if err := newNode.Connect(ctx, *pInfo); err != nil {
				logger.Info("bootstrapping a peer failed", err)
			}
			logger.Info("Parse Address:", addr)
		}
		//panic(err)
	}

	blockchain.Init()
	chain := blockchain.NewBlockchain(newNode, chainblockstore, bswap, configs)

	node.p2pNode = newNode
	node.mempool = mempool.NewMempool()
	node.pubsub = pubsub
	node.blockchain = chain
	node.wallet = wallet.NewWallet()

	node.ListenBlocks(ctx)
	node.ListenTransactions(ctx)

	return &node
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID.String() == n.h.ID().String() {
		return
	}
	logger.Info("discovered new peer ", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		logger.Info("error connecting to peer ", pi.ID.Pretty(), ": ", err)
	}
	logger.Info("connected to peer ", pi.ID.Pretty())
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

func (node *Node) ListenBlocks(ctx context.Context) {
	sub, err := node.pubsub.Subscribe("blocks")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				panic(err)
			}
			blk, err := block.DeserializeBlock(msg.GetData())
			if err != nil {
				panic(err)
			}
			// logger.Info("Block received over network:", string(blk.Serialize()))
			logger.Info("Block received over network, blockhash: ", blk.GetHash().String())
			cid := node.blockchain.AddBlock(blk)
			if cid != nil {
				logger.Info("Block added, cid:", cid)
				node.mempool.RemoveTxs(blk.Transactions)
			}
		}
	}()
}

func (node *Node) ListenTransactions(ctx context.Context) {
	sub, err := node.pubsub.Subscribe("transactions")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				panic(err)
			}
			tx, err := transaction.DeserializeTx(msg.GetData())
			if err != nil {
				panic(err)
			}
			node.mempool.AddTx(tx)
			logger.Info("Tx received over network, added to mempool:", tx)
		}
	}()
}

func (node *Node) CreateNewBlock() *block.Block {
	var blk block.Block
	height := node.blockchain.Head.Height + 1
	blkmsg, errmsg := block.ReadBlockMessage(height)
	if errmsg != nil {
		logger.Error(errmsg)
		return nil
	}
	//header
	blk.Header.PrevHash = node.blockchain.Head.GetHash()
	blk.Header.Version = "0.0.1"
	blk.Header.Timestamp = time.Now().Unix()
	blk.Header.Difficulty = node.blockchain.Head.Header.Difficulty
	blk.Header.Memo = blkmsg
	//body
	blk.Height = height
	blk.PrevCid = node.blockchain.GetBlockCid(node.blockchain.Head)
	blk.Reward = node.blockchain.CalcReward(blk.Height)
	blk.Transactions = node.mempool.SelectTransactions()
	blk.TxsCount = uint64(len(blk.Transactions))
	return &blk
}

func (node *Node) BroadcastBlock(block *block.Block) {
	data := block.Serialize()
	node.pubsub.Publish("blocks", data)
}

func (node *Node) GetBlock(height uint64) (*block.Block, error) {
	return node.blockchain.GetBlock(height)
}

func (node *Node) GetWallet() *wallet.Wallet {
	return node.wallet
}

func (node *Node) GetNewAddress() *NewAddressResponse {
	var res NewAddressResponse
	addr := node.wallet.GetNewAddress()
	res.Address = addr
	return &res
}

func (node *Node) SendTransaction(tx *transaction.Transaction) *SendTxResponse {
	// Check that node has key to send tx from address
	if node.wallet.GetStringAddress() == tx.From {
		var res SendTxResponse
		txid := tx.GetTxid()
		node.mempool.SetTransaction(txid, *tx)
		data := tx.Serialize()
		node.pubsub.Publish("transactions", data)
		res.Txid = tx.GetTxidString()
		return &res
	} else {
		logger.Info("Sending transaction failed")
		panic(errors.New("Sending tx failed, no key present in wallet"))
	}
}

func (node *Node) GetInfo() *GetInfoResponse {
	var res GetInfoResponse
	res.BlockHeight = node.blockchain.Head.Height
	return &res
}
