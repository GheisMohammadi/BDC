package transaction

import (
	hash "badcoin/src/helper/hash"
	number "badcoin/src/helper/number"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
)

type Transaction struct {
	ID      hash.Hash
	Inputs  []TXInput
	Outputs []TXOutput
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %v:", tx.ID))

	for i, input := range tx.Inputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:         %v", input.PreviousOutPoint.Hash.String()))
		lines = append(lines, fmt.Sprintf("       OutIndex:     %d", input.PreviousOutPoint.Index))
		lines = append(lines, fmt.Sprintf("       Signature:    %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:       %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:        %d", output.Amount))
		lines = append(lines, fmt.Sprintf("       PubKeyHash:   %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

func (tx *Transaction) Serialize() []byte {
	data, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return data
}

func DeserializeTx(buf []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(buf, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (tx *Transaction) GetTxid() hash.Hash {
	ser := tx.Serialize()
	hash, _ := hash.NewHash(ser)
	return *hash
}

func (tx *Transaction) GetTxidString() string {
	txid := tx.GetTxid()
	return hex.EncodeToString(txid[:])
}

// IsCoinBase checks whether the transaction is coinbase
func (tx Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].PreviousOutPoint.Hash.IsEqual(hash.ZeroHash()) && tx.Inputs[0].PreviousOutPoint.Index == -1
}

// Sign signs each input of a Transaction
// must match input's prev TX exists
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}

	//check input's prev TX exists
	for _, vin := range tx.Inputs {
		if _, exists := prevTXs[vin.PreviousOutPoint.StringHash()]; !exists {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	//get TX's trimmed copy
	txCopy := tx.TrimmedCopy()

	for inID, input := range txCopy.Inputs {
		prevTx := prevTXs[input.PreviousOutPoint.StringHash()]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[input.PreviousOutPoint.Index].PubKeyHash.CloneBytes() //why no use input's raw public key?

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inID].Signature = signature
		txCopy.Inputs[inID].PubKey = nil
	}
}

// Verify verifies signatures of Transaction inputs
// use signature & rawPubKey on ecdsa.Verify
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}

	for _, vin := range tx.Inputs {
		if _, exists := prevTXs[vin.PreviousOutPoint.StringHash()]; !exists {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Inputs {
		prevTx := prevTXs[vin.PreviousOutPoint.StringHash()]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[vin.PreviousOutPoint.Index].PubKeyHash.CloneBytes()

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Inputs[inID].PubKey = nil
	}

	return true
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
// set sign & pubkey nil
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Inputs {
		inputs = append(inputs, *NewTXInput(&vin.PreviousOutPoint, nil, nil))
	}
	for _, vout := range tx.Outputs {
		outputs = append(outputs, TXOutput{vout.Amount, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

// GetHash return the hash of the transaction
func (tx *Transaction) GetHash() *hash.Hash {
	return &tx.ID
}

// Hash calc and return the hash of the Transaction
func (tx *Transaction) CalcHash() hash.Hash {

	txCopy := *tx
	txCopy.ID = *hash.ZeroHash()

	h := hash.HashH(txCopy.Serialize())

	return h
}

func (tx *Transaction) StringHash() string {
	return tx.ID.String()
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string, reward int) *Transaction {
	if data == "" {
		data = number.GetRandData()
	}
	txin := NewTXInput(NewOutPoint(hash.ZeroHash(), -1), nil, []byte(data))
	txout := NewTXOutput(reward, to)
	tx := Transaction{*hash.ZeroHash(), []TXInput{*txin}, []TXOutput{*txout}}

	tx.ID = tx.CalcHash()

	return &tx
}

// NewUTXOTransaction creates a new transaction
// func NewUTXOTransaction(fromWallet *wallet.Wallet, to string, amount int, UTXOSet *UTXOSet, txPool TxPool) (*Transaction, error) {
// 	var inputs []TXInput
// 	var outputs []TXOutput

// 	pubKeyHash := wallet.HashPublicKey(fromWallet.PublicKey)
// 	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount, txPool)

// 	if acc < amount {
// 		return nil, errors.New("Not enough funds")
// 	}

// 	// Build a list of inputs
// 	for txid, outs := range validOutputs {
// 		var txID hash.Hash
// 		err := hash.Decode(&txID, txid)
// 		fmt.Println("NewUTXOTransaction", txID, txid)
// 		if err != nil {
// 			log.Panic(err)
// 		}

// 		for _, out := range outs {
// 			input := TXInput{*NewOutPoint(&txID, out), nil, fromWallet.PublicKey}
// 			inputs = append(inputs, input)
// 		}
// 	}

// 	// Build a list of outputs
// 	from := fromWallet.GetStringAddress()
// 	outputs = append(outputs, *NewTXOutput(amount, to))
// 	if acc > amount {
// 		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
// 	}

// 	tx := Transaction{*hash.ZeroHash(), inputs, outputs}
// 	tx.ID = tx.CalcHash()
// 	UTXOSet.Blockchain.SignTransaction(&tx, fromWallet.PrivateKey)

// 	//add TX to mempool
// 	_, err := txPool.MaybeAcceptTransaction(&tx)
// 	if err != nil {
// 		//TODO log err info
// 		logger.Error("NewUTXOTransaction error ",from," send ",to," ",amount," coins err: ",err)
// 		return nil, errors.New("add to mempool error: " + err.Error())
// 	}
// 	logger.Info("NewUTXOTransaction sucess ",from," send ",to," ",amount," coins tx: ",tx.StringHash())
// 	return &tx, nil
// }
