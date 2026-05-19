package hashing

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256(value string) string {
	h := sha256.New()
	h.Write([]byte(value))
	hashed := h.Sum(nil)
	hex := hex.EncodeToString(hashed)
	return hex
}
