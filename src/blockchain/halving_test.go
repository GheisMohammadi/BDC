package blockchain

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

func TestHalving(t *testing.T) {
	reward := CalcReward(210)
	fmt.Println(reward)
}

func CalcReward(height uint64) *big.Float {

	reward := new(big.Float)

	hcat := int64(height/100)
	halvingreward := 100 * math.Pow(0.5, float64(hcat))

	reward.SetFloat64(halvingreward)

	return reward
}
