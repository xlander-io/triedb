package triedb

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

type Hash [32]byte

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
func newHashFromBytes(input []byte) (result Hash) {
	if len(input) > 32 {
		copy(result[0:32], input[len(input)-32:])
		return
	} else {
		copy(result[32-len(input):], input)
		return
	}
}

func newHashFromString(input string) Hash {
	input_ := input
	if len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X') {
		input_ = input[2:]
	}
	if len(input)%2 == 1 {
		input_ = "0" + input_
	}

	h, _ := hex.DecodeString(input_)
	return newHashFromBytes(h)
}
