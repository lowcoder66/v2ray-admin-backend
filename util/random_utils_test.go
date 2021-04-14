package util

import (
	"fmt"
	"testing"
)

func TestRandomStr(t *testing.T) {
	count := 16

	fmt.Printf("random: %s\n", RandomStr(count))
	fmt.Printf("random: %s\n", RandomStr(count))
	fmt.Printf("random: %s\n", RandomStr(count))
}
