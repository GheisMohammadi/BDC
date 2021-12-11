package address

import (
	"bytes"
	"crypto/sha256"
	"log"

	"badcoin/src/helper/base58"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

// version pubkey for bitcoin, version = 0
const version = byte(0x00)
const addressChecksumLen = 4

// ToString returns address string format
func ToString(addr []byte) string {
	return fmt.Sprintf("%s", addr)
}

// GetAddress returns wallet address
// 1.hashes public key - ripemd160(sha256(public key))
// 2.connect version to the PubKeyHash header.
// 3.get checksum，use first 4 bytes
// 4.connect checksum to the end of the hash data
// 5.base58 data
func FromPublicKey(PublicKey []byte) []byte {
	//1.hashes public key - ripemd160(sha256(public key))
	pubKeyHash := HashPublicKey(PublicKey)
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
