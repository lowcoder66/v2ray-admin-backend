package util

import (
	"crypto/md5"
	"encoding/hex"
)

const salt = "v2ray-admin/2021"

func EncoderPassword(plain string) string {
	h := md5.New()
	h.Write([]byte(plain))
	h.Write([]byte(salt))

	st := h.Sum(nil)
	return hex.EncodeToString(st)
}

func PasswordMatch(plain string, cipher string) bool {
	return EncoderPassword(plain) == cipher
}
