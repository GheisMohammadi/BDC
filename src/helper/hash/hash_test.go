package hash

import (
	"testing"
	"fmt"
)

func TestHash_IsEqual(t *testing.T) {
	hOne := *ZeroHash()

	fmt.Println(hOne.IsEqual(ZeroHash()))

	fmt.Println(hOne)
	fmt.Println(ZeroHash())
}

func TestHashCid(t *testing.T) {
	hOne := *ZeroHash()

	fmt.Println(hOne.ToCid())

}