package blockchain

import (
	"crypto/ecdsa"
	"log"
	hash "badcoin/src/helper/hash"
	transaction "badcoin/src/transaction"
)

// SignTransaction signs inputs of a transaction.Transaction
func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Inputs {
		prevTX, err := bc.FindTransaction(&vin.PreviousOutPoint.Hash)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[prevTX.StringID()] = *prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID *hash.Hash) (*transaction.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}
		for _, tx := range block.Transactions {
			if tx.ID.IsEqual(ID) {
				return tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil, ErrorNotFoundTransaction
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}

	prevTXs := make(map[string]transaction.Transaction)
	for _, vin := range tx.Inputs {
		prevTX, err := bc.FindTransaction(&vin.PreviousOutPoint.Hash)
		if err != nil {
			logx.Error(err, vin.PreviousOutPoint.Hash.String())
			return false
		}
		prevTXs[prevTX.StringID()] = *prevTX
	}

	return tx.Verify(prevTXs)
}
