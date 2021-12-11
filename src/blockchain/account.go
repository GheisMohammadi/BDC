package blockchain

import (
	errors "badcoin/src/helper/error"
	"badcoin/src/transaction"
	"encoding/json"
	"math/big"
)

type Account struct {
	Nonce   uint64
	Address string
	Balance big.Int
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

func (chain *Blockchain) AddToAccountBalance(address string, value int64) error {
	bal := new(big.Int)
	bal.SetInt64(0)
	if accbytes, err := chain.Accounts.Get([]byte(address), nil); err != nil {
		return err
	} else {
		acc, _ := DeserializeAccount(accbytes)
		val := new(big.Int)
		val.SetInt64(value)
		res := new(big.Int).Add(&acc.Balance, val)
		if res.Cmp(big.NewInt(0)) == -1 {
			return errors.NotEnoughAccountBalance
		}
		acc.Balance.Set(res)
		//fmt.Println("acc balance updated:", acc.Balance.String())
		err = chain.StoreAccount(acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (chain *Blockchain) GetAccountBalance(address string) (*big.Int, error) {
	bal := new(big.Int)
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
	bal := new(big.Int)
	bal.SetInt64(0)
	if accbytes, err := chain.Accounts.Get([]byte(address), nil); err != nil {
		return nil, err
	} else {
		acc, _ := DeserializeAccount(accbytes)
		return acc, nil
	}
}

func (chain *Blockchain) UpdateAccounts(txs []transaction.Transaction) error {
	values := make(map[string]*big.Int)
	for _, tx := range txs {
		val := big.NewInt(int64(tx.Value))
		if values[tx.From] == nil {
			values[tx.From] = big.NewInt(0)
		}
		if values[tx.To] == nil {
			values[tx.To] = big.NewInt(0)
		}
		vf := values[tx.From]
		vt := values[tx.To]
		values[tx.To] = new(big.Int).Add(vt, val)
		values[tx.From] = new(big.Int).Add(vf, new(big.Int).Neg(val))
	}

	for _, val := range values {
		if val.Cmp(big.NewInt(0)) == -1 {
			return errors.NotEnoughAccountBalance
		}
	}
	return nil
}
