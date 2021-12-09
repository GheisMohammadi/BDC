package node

type HealthCheckResponse struct {
    Text     string
}   

type GetInfoResponse struct {
    BlockHeight     uint64
}   

type SendTxResponse struct {
    Txid       string
}

type NewAddressResponse struct {
    Address     string
}
