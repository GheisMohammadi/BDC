package blockchain

import (
	"log"

	//"github.com/michain/dotcoin/storage"
	block "badcoin/src/block"

	"github.com/ipfs/go-cid"
)

// Iterator implement a iterator for blockchain blocks
type Iterator struct {
	currentCid *cid.Cid
	bc         *Blockchain
}

// Next returns next block starting from the tip
func (i *Iterator) Next() *block.Block {

	blk, err := i.bc.LoadBlock(i.currentCid)
	if err != nil {
		log.Panic(err)
	}
	if blk != nil {
		i.currentCid = blk.PrevCid
	}
	return blk

}

// LocationHash locate current hash
func (i *Iterator) LocationHash(locateCid *cid.Cid) error {
	_, err := i.bc.LoadBlock(locateCid)
	if err == nil {
		i.currentCid = locateCid
	}
	return err
}
