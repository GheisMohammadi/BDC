package mempool

import (
	hash "badcoin/src/helper/hash"
	"badcoin/src/transaction"
	"math/big"
	logger "badcoin/src/helper/logger"
)

type getAcc func(addr string) (*big.Float, uint64, error)
type Mempool struct {
	transactions map[hash.Hash]transaction.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[hash.Hash]transaction.Transaction),
	}
}

func (mempool *Mempool) AddTx(tx *transaction.Transaction) {
	txid := tx.GetTxid()
	mempool.transactions[txid] = *tx
}

func (mempool *Mempool) RemoveTxs(txs []*transaction.Transaction) {
	for _, tx := range txs {
		txid := tx.GetTxid()
		delete(mempool.transactions, txid)
	}
}

func (mempool *Mempool) SelectTransactions(f getAcc) []*transaction.Transaction {
	var txs []*transaction.Transaction
	for _, tx := range mempool.transactions {
		addr := tx.From
		if bal, nonce, err := f(addr); err != nil {
			return make([]*transaction.Transaction, 0)
		} else {
            //value should be less than balance and also checking the nonce 
			if bal.Cmp(big.NewFloat(tx.Value)) >= 0 && tx.Nonce == nonce+1 {
				txs = append(txs, &tx)
			} else {
				logger.Info("tx with value:",tx.Value,"rejected from mempool. acc balance is: ", bal.String()," nonce: ",tx.Nonce," and account nonce is: ",nonce)
			}
		}
	}
	return txs
}

func (mempool *Mempool) SetTransaction(txid hash.Hash, tx transaction.Transaction) {
	mempool.transactions[txid] = tx
}

func (mempool *Mempool) TransactionsCount() int {
	return len(mempool.transactions)
}

func (mempool *Mempool) Clear() {
	mempool.transactions = make(map[hash.Hash]transaction.Transaction)
}
