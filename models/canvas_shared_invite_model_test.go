package model

import (
	"testing"
)

func TestGenerateRandomCode(t *testing.T) {
	lengths := []uint{0, 1, 20, 200, 550, 2048, 99999}

	for _, l := range lengths {
		code, err := generateCode(l)
		if err != nil {
			codeLen := len(code)

			if codeLen != int(l) {
				t.Errorf("Expected code of length: %v, got length: %v", l, codeLen)
			}
		}
	}
}
