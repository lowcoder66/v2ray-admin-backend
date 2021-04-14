package util

import (
	"fmt"
	"testing"
)

func TestEncoder(t *testing.T) {
	plain := "?123456"
	cipher := EncoderPassword(plain)

	fmt.Printf("plain:%s cipher:%s\n", plain, cipher)
	if !PasswordMatch(plain, cipher) {
		t.Error("error")
	}
}
