package util

import (
	"bytes"
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

type Hash [32]byte

var NIL_HASH = NewHashFromString("0xa7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a")

func IsEmptyHash(hash *Hash) bool {
	if hash == nil {
		return true
	}
	return bytes.Compare((*hash)[:], NIL_HASH[:]) == 0
}

// sha3-256 hash
// todo check sha3_g.Write error exist or not?
func hash(input []byte) []byte {
	sha3_g := sha3.New256()
	// Create a new hash & write input string
	_, _ = sha3_g.Write([]byte(input))
	// Get the resulting encoded byte slice
	return sha3_g.Sum(nil)
}

// case 1 , input is nil , return const of empty hash
// case 2 , input is longer then 32 , return the first 32 bytes of hash
func NewHashFromBytes(input []byte) *Hash {
	var result Hash
	if len(input) > 32 {
		copy(result[0:32], input[len(input)-32:])
	} else {
		copy(result[32-len(input):], input)
	}
	return &result
}

func NewHashFromString(input string) *Hash {
	input_ := input
	if len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X') {
		input_ = input[2:]
	}
	if len(input)%2 == 1 {
		input_ = "0" + input_
	}

	h, _ := hex.DecodeString(input_)
	return NewHashFromBytes(h)
}
