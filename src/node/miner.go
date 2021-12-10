package node

import (
	"encoding/base64"
	"math/rand"
	"time"

	// "math/big"
	block "badcoin/src/block"
	logger "badcoin/src/helper/logger"
)

func (node *Node) StartMiner() {
	c := make(chan *block.Block)
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

func isWinner(ticket string) bool {
	// num := big.NewInt(0).SetBytes(ticket)
	winner := "000"
	res := false
	if ticket[:len(winner)] == winner {
		return true
	}
	return res
}

func (node *Node) FindSolsHash(c chan *block.Block) {
	blk := node.CreateNewBlock()
	blk.Header.Nonce = make([]byte, 32)
	for {
		rand.Read(blk.Header.Nonce)
		guess := blk.CalcHash()
		ticket := base64.StdEncoding.EncodeToString(guess[:])
		if isWinner(ticket) {
			blk.Header.Solution = ticket
			// logger.Info("Ticket:", ticket)
			c <- blk
			blk = node.CreateNewBlock()
			blk.Header.Nonce = make([]byte, 32)
		}
	}
}
