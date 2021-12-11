package transaction

import (
	"testing"
)

func TestBlockchain(t *testing.T) {
	tx := NewCoinbaseTX("ABCDEF","Test Tx",100)
	t.Log(tx.String())
}
