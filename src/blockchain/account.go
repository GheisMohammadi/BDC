package blockchain

import (
	errors "badcoin/src/helper/error"
	logger "badcoin/src/helper/logger"
	"badcoin/src/transaction"
	"encoding/json"
	"math/big"

	"github.com/syndtr/goleveldb/leveldb"
)

type Account struct {
	Nonce   uint64
	Address string
	Balance big.Float
}

func (acc *Account) Serialize() []byte {
	data, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}
	return data
}

func DeserializeAccount(buf []byte) (*Account, error) {
	var acc Account
	err := json.Unmarshal(buf, &acc)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (chain *Blockchain) StoreAccount(acc *Account) error {
	data := acc.Serialize()
	return chain.Accounts.Put([]byte(acc.Address), data, nil)
}

func (chain *Blockchain) AddToAccountBalance(address string, value float64, increasenonce bool) error {
	bal := new(big.Float)
	bal.SetInt64(0)
	var acc *Account
	if accbytes, err := chain.Accounts.Get([]byte(address), nil); err != nil {
		if err != leveldb.ErrNotFound {
			return err
		}
		acc = new(Account)
		acc.Address = address
		acc.Balance = *big.NewFloat(0)
		acc.Nonce = uint64(0)
	} else {
		acc, _ = DeserializeAccount(accbytes)
	}

	val := new(big.Float)
	val.SetFloat64(value)
	res := new(big.Float).Add(&acc.Balance, val)
	if res.Cmp(big.NewFloat(0)) == -1 {
		return errors.NotEnoughAccountBalance
	}
	acc.Balance.Set(res)
	if increasenonce {
		acc.Nonce++
	}
	logger.Info("acc balance updated:", acc.Balance.String())
	err := chain.StoreAccount(acc)
	if err != nil {
		return err
	}

	return nil
}

func (chain *Blockchain) GetAccountBalance(address string) (*big.Float, error) {
	bal := new(big.Float)
	bal.SetInt64(0)
	if accbytes, err := chain.Accounts.Get([]byte(address), nil); err != nil {
		return nil, err
	} else {
		acc, _ := DeserializeAccount(accbytes)
		bal.Set(&acc.Balance)
	}
	return bal, nil
}

func (chain *Blockchain) FetchAccountDetails(address string) (*Account, error) {
	bal := new(big.Float)
	bal.SetInt64(0)
	if accbytes, err := chain.Accounts.Get([]byte(address), nil); err != nil {
		return nil, err
	} else {
		acc, _ := DeserializeAccount(accbytes)
		return acc, nil
	}
}

func (chain *Blockchain) CalcAccountsUpdates(txs []*transaction.Transaction) (map[string]*big.Float,error) {
	values := make(map[string]*big.Float)
	for _, tx := range txs {
		val := big.NewFloat(float64(tx.Value))
		if values[tx.From] == nil {
			values[tx.From] = big.NewFloat(0)
		}
		if values[tx.To] == nil {
			values[tx.To] = big.NewFloat(0)
		}
		vf := values[tx.From]
		vt := values[tx.To]
		values[tx.To] = new(big.Float).Add(vt, val)
		values[tx.From] = new(big.Float).Add(vf, new(big.Float).Neg(val))
	}

	for addr, val := range values {
		bal, err := chain.GetAccountBalance(addr)
		if err != nil {
			if err != leveldb.ErrNotFound {
				return nil,errors.CheckAccountBalanceFailed
			}
			bal = big.NewFloat(0)
		}
		newbal := big.NewFloat(0).Add(bal, val)
		if newbal.Cmp(big.NewFloat(0)) == -1 {
			return nil,errors.NotEnoughAccountBalance
		}
	}

	for addr, val := range values {
		bal, err := chain.GetAccountBalance(addr)
		if err != nil {
			if err != leveldb.ErrNotFound {
				return nil,errors.CheckAccountBalanceFailed
			}
			bal = big.NewFloat(0)
		}
		newbal := big.NewFloat(0).Add(bal, val)
		if newbal.Cmp(big.NewFloat(0)) == -1 {
			return nil,errors.NotEnoughAccountBalance
		}
	}

	return values,nil
}

func (chain *Blockchain) UpdateAccounts(values map[string]*big.Float) error {

	for addr, val := range values {
		addvalue,_ := val.Float64()
		chain.AddToAccountBalance(addr,addvalue,true)
	}

	return nil
}
