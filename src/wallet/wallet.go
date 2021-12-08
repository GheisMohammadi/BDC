package wallet 

import (
    "crypto/rand"
    "crypto/ecdsa"
    "crypto/elliptic"
	"badcoin/src/helper/hash"
    "encoding/base64"
)

type Wallet struct {
    // super insecure keydump, address:privkey
    keyDump map[string]ecdsa.PrivateKey 
}

func NewWallet() *Wallet{
    return &Wallet{
        keyDump: make(map[string]ecdsa.PrivateKey),
    }
}

func NewKey() *ecdsa.PrivateKey {
    privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        panic(err)
    }
	return privKey
}

func ToAddress(pubkey ecdsa.PublicKey) string {
    pubBytes := elliptic.Marshal(elliptic.P256(), pubkey.X, pubkey.Y)
    addr := hash.HashB(pubBytes[1:])
    addrString := base64.StdEncoding.EncodeToString(addr[:])[12:]
    return addrString
}

func (wallet *Wallet) GetNewAddress() string {
    new := NewKey()
    addr := ToAddress(new.PublicKey)
    wallet.keyDump[addr] = *new
    return addr
}

func (wallet *Wallet) HasKey(addr string) bool {
    _, ok := wallet.keyDump[addr]
    return ok
}
