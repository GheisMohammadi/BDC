package blockchain

import (
	"log"

	hash "badcoin/src/helper/hash"

	//"github.com/michain/dotcoin/storage"
	block "badcoin/src/block"
)

// Iterator implement a iterator for blockchain blocks
type Iterator struct {
	currentHash hash.Hash
	bc          *Blockchain
}

// Next returns next block starting from the tip
func (i *Iterator) Next() *block.Block {

	blk, err := i.bc.LoadBlock(&i.currentHash)
	if err != nil {
		log.Panic(err)
	}
	if blk != nil {
		i.currentHash = blk.Header.PrevHash
	}
	return blk

}

// LocationHash locate current hash
func (i *Iterator) LocationHash(locateHash *hash.Hash) error {
	_, err := i.bc.LoadBlock(locateHash)
	if err == nil {
		i.currentHash = *locateHash
	}
	return err
}
