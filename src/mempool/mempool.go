package mempool 

import (
    "badcoin/src/transaction"
    hash "badcoin/src/helper/hash"
)
type Mempool struct {
    transactions map[hash.Hash]transaction.Transaction
}

func NewMempool() *Mempool{
    return &Mempool{
        transactions: make(map[hash.Hash]transaction.Transaction),
    }
}

func (mempool *Mempool) AddTx(tx *transaction.Transaction) {
    txid := tx.GetTxid()
    mempool.transactions[txid] = *tx
}

func (mempool *Mempool) RemoveTxs(txs []transaction.Transaction){
    for _, tx := range txs {
        txid := tx.GetTxid()
        delete(mempool.transactions, txid)
    }
}

func (mempool *Mempool) SelectTransactions() []transaction.Transaction {
    var txs []transaction.Transaction
    for _, v := range mempool.transactions {
        txs = append(txs, v)
    }
    return txs
}

func (mempool *Mempool) SetTransaction(txid hash.Hash, tx transaction.Transaction) {
    mempool.transactions[txid] = tx
}
