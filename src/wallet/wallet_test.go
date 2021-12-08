package wallet 

import (
    "fmt"
    "testing"
)

func TestToAddress(t *testing.T){
    new := NewKey()
    addr := ToAddress(new.PublicKey)
    fmt.Println("Address:", addr)
}