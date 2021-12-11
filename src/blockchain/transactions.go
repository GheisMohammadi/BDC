package blockchain

import (
	errors "badcoin/src/helper/error"
	hash "badcoin/src/helper/hash"
	transaction "badcoin/src/transaction"
)

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID *hash.Hash) (*transaction.Transaction, error) {
	bci := bc.GetIterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}
		for _, tx := range block.Transactions {
			if tx.ID.IsEqual(ID) {
				return &tx, nil
			}
		}

		if len(block.Header.PrevHash) == 0 {
			break
		}
	}

	return nil, errors.NotFoundTransaction
}
