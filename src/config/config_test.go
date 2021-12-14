package config

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	configs, _ := Init("")
	fmt.Println(configs.ID)
}
