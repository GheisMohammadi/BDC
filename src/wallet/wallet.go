package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"badcoin/src/helper/base58"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

// version pubkey for bitcoin, version = 0
const version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
	Nonce      uint64
}

// newWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public, uint64(0)}

	return &wallet
}

// GetStringAddress returns wallet address string format
func (w Wallet) GetStringAddress() string {
	return fmt.Sprintf("%s", w.GetAddress())
}

// GetAddress returns wallet address
// 1.hashes public key - ripemd160(sha256(public key))
// 2.connect version to the PubKeyHash header.
// 3.get checksum，use first 4 bytes
// 4.connect checksum to the end of the hash data
// 5.base58 data
func (w Wallet) GetAddress() []byte {
	//1.hashes public key - ripemd160(sha256(public key))
	pubKeyHash := HashPublicKey(w.PublicKey)
	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := base58.Encode(fullPayload)

	return []byte(address)
}

// HashPublicKey hashes public key
// 1.sha256 publick key
// 2.ripemd160(sha256(public key))
func HashPublicKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress check if address if valid
// 1.base58 decode
// 2.get checksum(4 byte)
// 3.get version
// 4.get pubKeyHash
// 5.get checksum，use first 4 bytes
func ValidateAddress(address string) bool {
	pubKeyHash := base58.Decode(address)
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// GetPubKeyHashFromAddress get PublicKeyHash from address
// 1.base58 decode
// 2.remove version & checksum
func GetPubKeyHashFromAddress(address []byte) []byte {
	pubKeyHash := base58.Decode(string(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	return pubKeyHash
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

// newKeyPair create private&public key with ecdsa and rand-key
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

func (wallet *Wallet) GetNewAddress() string {
	prv, pub := newKeyPair()
	wallet.PrivateKey = prv
	wallet.PublicKey = pub
	wallet.Nonce = 0
	return wallet.GetStringAddress()
}

func (wallet *Wallet) AddNonce() uint64 {
	wallet.Nonce++
	return wallet.Nonce
}

func (wallet *Wallet) SetNonce(nonce uint64) {
	wallet.Nonce = nonce
}