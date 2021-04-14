package util

import (
	"math/rand"
)

var numbers, letters = []rune("0123456789"), []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomStr(count int) string {
	chars := append(numbers, letters...)

	b := make([]rune, count)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}
