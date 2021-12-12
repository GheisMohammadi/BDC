package node

import (
	"context"
	"errors"
	"os"
	"time"

	block "badcoin/src/block"
	blockchain "badcoin/src/blockchain"
	logger "badcoin/src/helper/logger"
	mempool "badcoin/src/mempool"
	transaction "badcoin/src/transaction"
	wallet "badcoin/src/wallet"

	config "badcoin/src/config"

	proofofwork "badcoin/src/pow"

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
	pow        *proofofwork.ProofOfWork
}

func CreateNewNode(ctx context.Context, configs *config.Configurations) *Node {
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
		logger.Info(i, ": ", addr.String() + "/ipfs/" + newNode.ID().Pretty())
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
	blockchain := blockchain.NewBlockchain(newNode, configs)

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
