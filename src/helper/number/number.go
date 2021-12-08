package number

import (
	"badcoin/src/helper/uuid"
	"bytes"
	"encoding/binary"
	"log"
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
