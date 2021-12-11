package transaction

import (
	hash "badcoin/src/helper/hash"
	"bytes"
	"encoding/gob"
	"log"

	wallet "badcoin/src/wallet"
)

// TXOutput represents a transaction output
type TXOutput struct {
	Amount     int
	PubKeyHash hash.Hash
}

// Lock set PublicKeyHash to signs the output
// input must check this value to use
func (out *TXOutput) Lock(address []byte) {
	keybytes := wallet.GetPubKeyHashFromAddress(address)
	keyhash, _ := hash.NewHash(keybytes)
	out.PubKeyHash = *keyhash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash hash.Hash) bool {
	return out.PubKeyHash == pubKeyHash
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	out := &TXOutput{value, *hash.ZeroHash()}
	out.Lock([]byte(address))
	return out
}

// Serialize serializes []TXOutput
func SerializeOutputs(outs []TXOutput) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) []TXOutput {
	var outputs []TXOutput

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
