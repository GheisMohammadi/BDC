package number

import (
	"badcoin/src/helper/uuid"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
)

func GetRandData() string {
	return uuid.NewV4().String()
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

//Int64ToByteArray is second way to convert int64 to byte array
func Int64ToByteArray(num int64) []byte {
	numbig := new(big.Int)
	numbig.SetInt64(num)
	numbytes := numbig.Bytes()
	return numbytes
}

func RoundBigFloat(x *big.Float) *big.Float {
	newdec := fmt.Sprintf("%.8f\n", x)
	res := new(big.Float)
	res.SetString(newdec)
	return res
}
