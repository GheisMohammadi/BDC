package node

import (
	"math/big"
)

type HealthCheckResponse struct {
	Text string
}

type GetInfoResponse struct {
	BlockHeight uint64
	NodeAddress string
	NodeBalance *big.Float
}

type SendTxResponse struct {
	Txid string
}

type NewAddressResponse struct {
	Address string
}
