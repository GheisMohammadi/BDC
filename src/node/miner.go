package node

import (
	"math/rand"
	"time"

	// "math/big"
	block "badcoin/src/block"
	logger "badcoin/src/helper/logger"
	merkle "badcoin/src/merkle"
	proofofwork "badcoin/src/pow"
)

func (node *Node) StartMiner() {
	c := make(chan *block.Block)
	node.pow = proofofwork.NewProofOfWorkT(1)
	go node.Mine(c)
}

func (node *Node) Mine(c chan *block.Block) {
	go node.FindSolsHash(c)
	for {
		select {
		case blk := <-c:
			node.BroadcastBlock(blk)
		}
	}
}

func FindSolsTimeout(c chan string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		rand := r.Intn(10) // Adjusts variance of block speed
		logger.Info("Interval:", rand)
		time.Sleep(time.Duration(rand) * time.Second)
		c <- "Found a block solution"
	}
}

func (node *Node) FindSolsHash(c chan *block.Block) {

	blk := node.CreateNewBlock()
	logger.Info("Mining started!")
	//blk.Header.Nonce = make([]byte, 32)

	for {
		mtree := merkle.BuildTxMerkleTree(blk.Transactions)
		rootHash := mtree.RootNode.Data
		//fmt.Println(rootHash)
		difficulty := node.blockchain.AdjustDifficulty(blk)
		node.pow.SetTarget(int(difficulty))

		solved := node.pow.SolveHash(blk.Header.PrevHash[:], rootHash, nil)
		if solved == true {
			//blk.Header.Solution = node.pow.Hash.String()
			blk.Header.Nonce = node.pow.Nonce
			now := time.Now()
			blk.Header.Timestamp = now.UnixMicro()
			blk.UpdateHash()
			c <- blk
			blkstr := string(blk.Serialize())
			logger.Info("Block #", blk.Height, ": ", blkstr)
			//blk.Header.Nonce = make([]byte, 32)
		}
		time.Sleep(time.Duration(10) * time.Second)
		blk = node.CreateNewBlock()
	}

}
