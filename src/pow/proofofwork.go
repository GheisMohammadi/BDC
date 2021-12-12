package pow

import (
	hash "badcoin/src/helper/hash"
	logger "badcoin/src/helper/logger"
	"badcoin/src/helper/number"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"math/big"
	"time"
)

var (
	maxNonce = math.MaxInt64
)

const TargetBits = 16

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	Target   *big.Int
	Nonce    int64
	Hash     hash.Hash
	Duration int64
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWorkT(targetBits int) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{Target: target}
	return pow
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork() *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-TargetBits))
	pow := &ProofOfWork{Target: target}
	return pow
}

// NewProofOfWork builds and returns a ProofOfWork
func (pow *ProofOfWork) SetTarget(targetBits int) error {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-TargetBits))
	pow.Target.Set(target)
	return nil
}

// calculateHash calc hash with bestBlockHash and Txs hashes
func (pow *ProofOfWork) calculateHash(prevBlockHash, TXsHash []byte, data []byte, nonce int) [32]byte {
	datatohash := bytes.Join(
		[][]byte{
			prevBlockHash,
			TXsHash,
			data,
			number.IntToHex(int64(TargetBits)),
			number.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return sha256.Sum256(datatohash)
}

// solveHash solve right hash which less than the target difficulty
// it will be stop when received quit signal
func (pow *ProofOfWork) solveHash(prevBlockHash, TXsHash []byte, data []byte, quit chan struct{}) bool {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	t1 := time.Now().UnixMicro()
	for nonce < maxNonce {
		select {
		case <-quit:
			logger.Trace("Mining SolveHash Failed, because receive quit signal")
			return false
		default:
			hash = pow.calculateHash(prevBlockHash, TXsHash, data, nonce)

			// if math.Remainder(float64(nonce), 10000) == 0 {
			// 	//fmt.Printf("\r%x", hash)
			// }

			hashInt.SetBytes(hash[:])
			if hashInt.Cmp(pow.Target) == -1 {
				t2 := time.Now().UnixMicro()
				pow.Duration = t2 - t1
				pow.Nonce = int64(nonce)
				pow.Hash.SetBytes(hash[:])
				logger.Trace("Mining SolveHash Success", nonce, hex.EncodeToString(hash[:]))
				return true
			} else {
				nonce++
			}
		}
	}
	logger.Trace("Mining SolveHash Failed, nonce now is same to maxNonce", nonce, maxNonce)
	return false
}

func (pow *ProofOfWork) RunAtOnce(prevBlockHash, TXsHash []byte, data []byte) (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	//immediately return for test
	hash = pow.calculateHash(prevBlockHash, TXsHash, data, nonce)
	hashInt.SetBytes(hash[:])
	return nonce, hash[:]
}

// SolveHash loop calc hash to solve target
func (pow *ProofOfWork) SolveHash(prevBlockHash, TXsHash []byte, data []byte, quit chan struct{}) bool {
	isSolve := pow.solveHash(prevBlockHash, TXsHash, data, quit)
	return isSolve
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate(prevBlockHash, TXsHash []byte, data []byte, nonce int) bool {
	var hashInt big.Int

	hash := pow.calculateHash(prevBlockHash, TXsHash, data, nonce)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.Target) == -1

	return isValid
}
