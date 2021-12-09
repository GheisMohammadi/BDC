package node

import (
	"context"
	"errors"
	"os"
	"time"

	block "badcoin/src/block"
	blockchain "badcoin/src/blockchain"
	"badcoin/src/helper/hash"
	logger "badcoin/src/helper/logger"
	mempool "badcoin/src/mempool"
	transaction "badcoin/src/transaction"
	wallet "badcoin/src/wallet"

	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	floodsub "github.com/libp2p/go-libp2p-pubsub"
)

type Node struct {
	p2pNode    host.Host
	mempool    *mempool.Mempool
	blockchain *blockchain.Blockchain
	pubsub     *floodsub.PubSub
	wallet     *wallet.Wallet
}

func CreateNewNode(ctx context.Context) *Node {
	var node Node

	newNode, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		panic(err)
	}

	pubsub, err := floodsub.NewFloodSub(ctx, newNode)
	if err != nil {
		panic(err)
	}

	for i, addr := range newNode.Addrs() {
		logger.Info(i,": ",addr,"/ipfs/",newNode.ID().Pretty())
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

	blockchain := blockchain.NewBlockchain(newNode)

	node.p2pNode = newNode
	node.mempool = mempool.NewMempool()
	node.pubsub = pubsub
	node.blockchain = blockchain
	node.wallet = wallet.NewWallet()

	node.ListenBlocks(ctx)
	node.ListenTransactions(ctx)

	return &node
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
			logger.Info("Block received over network, blockhash", blk.GetCid())
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
	cid := node.blockchain.Head.GetCid()
	h,_ := hash.FromCid(&cid)
	blk.PrevHash = *h
	blk.Transactions = node.mempool.SelectTransactions()
	blk.Height = node.blockchain.Head.Height + 1
	blk.Timestamp = uint64(time.Now().Unix())
	return &blk
}

func (node *Node) BroadcastBlock(block *block.Block) {
	data := block.Serialize()
	node.pubsub.Publish("blocks", data)
}

func (node *Node) GetNewAddress() *NewAddressResponse {
	var res NewAddressResponse
	addr := node.wallet.GetNewAddress()
	res.Address = addr
	return &res
}

func (node *Node) SendTransaction(tx *transaction.Transaction) *SendTxResponse {
	// Check that node has key to send tx from address
	if node.wallet.HasKey(tx.Sender) {
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
