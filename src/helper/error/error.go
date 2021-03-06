package error

import (
	"errors"
)

var BlockNotFount = errors.New("Block is not found")

var StorageInitFailed = errors.New("storage initialization failed")

var InvalidHash = errors.New("Invalid hash")

var InvalidHeight = errors.New("Invalid block height")

var NotFoundTransaction = errors.New("not found the transaction")

var BlockNoTransactions = errors.New("block does not contain any transactions")

var BlockSizeTooBig = errors.New("block serialized is too big")

var BlockTooManyTransactions = errors.New("block has too many transactions")

var BlockBadMerkleRoot = errors.New("block merkle root is invalid")

var BlockDuplicateTx = errors.New("block contains duplicate transaction")

var NotEnoughAccountBalance = errors.New("Not enough account balance")

var CheckAccountBalanceFailed = errors.New("checking of account balance failed")

var ExchangeISNotOnline = errors.New("Exchange is not online")

var InvalidNonce = errors.New("Nonce is invalid")

var AlreadyHasPendingTx = errors.New("Account already has pending transaction")
