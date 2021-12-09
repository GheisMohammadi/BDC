package transaction

import (
	hash "badcoin/src/helper/hash"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type Transaction struct {
	ID       hash.Hash
	Sender   string
	Receiver string
	Amount   uint64
	Memo     string
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %v:", tx.ID))

	// for i, input := range tx.Inputs {

	// 	lines = append(lines, fmt.Sprintf("     Input %d:", i))
	// 	lines = append(lines, fmt.Sprintf("       TXID:         %v", input.PreviousOutPoint.Hash.String()))
	// 	lines = append(lines, fmt.Sprintf("       OutIndex:     %d", input.PreviousOutPoint.Index))
	// 	lines = append(lines, fmt.Sprintf("       Signature:    %x", input.Signature))
	// 	lines = append(lines, fmt.Sprintf("       PubKey:       %x", input.PubKey))
	// }

	// for i, output := range tx.Outputs {
	// 	lines = append(lines, fmt.Sprintf("     Output %d:", i))
	// 	lines = append(lines, fmt.Sprintf("       Value:        %d", output.Value))
	// 	lines = append(lines, fmt.Sprintf("       PubKeyHash:   %x", output.PubKeyHash))
	// }

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
