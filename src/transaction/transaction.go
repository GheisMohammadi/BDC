package transaction

import (
	address "badcoin/src/helper/address"
	hash "badcoin/src/helper/hash"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
)

type Transaction struct {
	ID        hash.Hash
	PublicKey []byte
	Signature []byte
	Timestamp int64
	From      string
	To        string
	Fee       uint64
	Value     uint64
	Data      string
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %v:", tx.ID))
	lines = append(lines, fmt.Sprintf("       ID:           %v", tx.ID.String()))
	lines = append(lines, fmt.Sprintf("       Time:         %d", tx.Timestamp))
	lines = append(lines, fmt.Sprintf("       From:         %s", tx.From))
	lines = append(lines, fmt.Sprintf("       To:           %s", tx.To))
	lines = append(lines, fmt.Sprintf("       PublicKey:    %v", tx.PublicKey))
	lines = append(lines, fmt.Sprintf("       Fee:		    %d", tx.Fee))
	lines = append(lines, fmt.Sprintf("       Value:		%d", tx.Value))
	lines = append(lines, fmt.Sprintf("       Signature:    %x", tx.Signature))
	lines = append(lines, fmt.Sprintf("       Data:         %x", tx.Data))

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

func NewTransaction(pubKey []byte, to string, value uint64, data string) *Transaction {

	fromBytes := address.FromPublicKey(pubKey)
	from := address.ToString(fromBytes)

	now := time.Now()

	tx := Transaction{
		ID:        *hash.ZeroHash(),
		PublicKey: pubKey,
		Signature: []byte{},
		Timestamp: now.UnixMilli(),
		From:      from,
		To:        to,
		Fee:       0,
		Value:     value,
		Data:      data,
	}

	tx.UpdateHash()

	return &tx
}

func (tx *Transaction) GetTxid() hash.Hash {
	tx.ID = *hash.ZeroHash()
	ser := tx.Serialize()
	hash, _ := hash.NewHash(ser)
	return *hash
}

func (tx *Transaction) GetTxidString() string {
	txid := tx.GetTxid()
	return txid.String()
}

// Sign signs each input of a Transaction
// must match input's prev TX exists
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey) {

	txCopy := tx.TrimmedCopyToSign()
	dataToSign := fmt.Sprintf("%x\n", txCopy)
	r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(dataToSign))
	if err != nil {
		log.Panic(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature

}

// Verify verifies signatures of Transaction inputs
// use signature & rawPubKey on ecdsa.Verify
func (tx *Transaction) VerifySignature() bool {

	txCopy := tx.TrimmedCopyToSign()
	curve := elliptic.P256()

	r := big.Int{}
	s := big.Int{}
	sigLen := len(tx.Signature)
	r.SetBytes(tx.Signature[:(sigLen / 2)])
	s.SetBytes(tx.Signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(tx.PublicKey)
	x.SetBytes(tx.PublicKey[:(keyLen / 2)])
	y.SetBytes(tx.PublicKey[(keyLen / 2):])

	dataToVerify := fmt.Sprintf("%x\n", txCopy)

	rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
	if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
		return false
	}

	return true
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
// set sign & pubkey nil
func (tx *Transaction) TrimmedCopyToSign() Transaction {
	txCopy := Transaction{*hash.ZeroHash(), []byte{}, []byte{}, tx.Timestamp, tx.From, tx.To, tx.Fee, tx.Value, tx.Data}
	return txCopy
}

func (tx *Transaction) TrimmedCopy() Transaction {
	txCopy := Transaction{tx.ID, tx.PublicKey, tx.Signature, tx.Timestamp, tx.From, tx.To, tx.Fee, tx.Value, tx.Data}
	return txCopy
}

// GetHash return the hash of the transaction
func (tx *Transaction) GetHash() *hash.Hash {
	return &tx.ID
}

// GetHash return the hash of the transaction
func (tx *Transaction) UpdateHash() error {
	txHash := tx.CalcHash()
	tx.ID = txHash
	return nil
}

// Hash calc and return the hash of the Transaction
func (tx *Transaction) CalcHash() hash.Hash {
	txCopy := tx.TrimmedCopyToSign()
	h := hash.HashH(txCopy.Serialize())
	return h
}

func (tx *Transaction) StringHash() string {
	return tx.ID.String()
}
