package node

import (
	"time"

	// "math/big"
	block "badcoin/src/block"
	config "badcoin/src/config"
	hash "badcoin/src/helper/hash"
	logger "badcoin/src/helper/logger"
	merkle "badcoin/src/merkle"
	proofofwork "badcoin/src/pow"
)

func (node *Node) StartMiner(configs *config.Configurations) {
	c := make(chan *block.Block)
	node.pow = proofofwork.NewProofOfWorkT(1)
	go node.Mine(c, configs.Mining.ExpectedMiningTimeInSeconds)
}

func (node *Node) Mine(c chan *block.Block, blocktime uint64) {
	go node.FindSolsHash(c, blocktime)
	for {
		select {
		case blk := <-c:
			//node.blockchain.PutBlock(blk)
			node.BroadcastBlock(blk)
		}
	}
}

func (node *Node) FindSolsHash(c chan *block.Block, expectedblocktime uint64) {

	blk := node.CreateNewBlock()
	if blk == nil {
		logger.Error("Can't create new block")
		return
	}
	logger.Info("Mining started!")
	//blk.Header.Nonce = make([]byte, 32)

	for {
		mtree := merkle.BuildTxMerkleTree(blk.Transactions)
		rootData := mtree.RootNode.Data
		rootHash, _ := hash.FromByteArray(rootData)
		blk.Header.MerkleRoot = *rootHash
		//fmt.Println(rootHash)
		difficulty := node.blockchain.AdjustDifficulty(blk)
		node.pow.SetTarget(int(difficulty))

		extradata := []byte(blk.Header.Miner)
		solved := node.pow.SolveHash(blk.Header.PrevHash[:], rootHash.CloneBytes(), extradata, nil)
		if solved == true {
			//blk.Header.Solution = node.pow.Hash.String()
			blk.Header.Nonce = node.pow.Nonce
			now := time.Now()
			blk.Header.Timestamp = now.UnixMicro()
			blk.Header.Miner = node.wallet.GetStringAddress()
			blk.UpdateHash()
			c <- blk
			blkstr := string(blk.Serialize())
			logger.Info("Block #", blk.Height, ": ", blkstr)
			//blk.Header.Nonce = make([]byte, 32)
		}
		time.Sleep(time.Duration(expectedblocktime) * time.Second)
		blk = node.CreateNewBlock()
		if blk == nil {
			logger.Error("Can't create new block, Mining will be stopped")
			return
		}
	}

}
