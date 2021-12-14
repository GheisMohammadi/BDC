package file

import (
	"fmt"
	"testing"
)

func TestFile(t *testing.T) {
	existed := IsExist("file.go")
	fmt.Println(existed)
	if existed == false {
		t.Error("exist checking failed")
	}
}
