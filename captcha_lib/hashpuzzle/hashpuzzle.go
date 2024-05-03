package hashpuzzle

import (
	"crypto/sha256"
	"fmt"
)

// HashNb generates a hash puzzle based on the input byte string and a salt
// It requires N >= 0
func HashNb(bs []byte, N uint16, salt []byte) []byte {
	bsr := bs
	var h = sha256.New()
	h.Write(bsr)
	h.Write(salt)
	bsr = h.Sum(nil)
	for i := uint16(1); i < N; i++ {
		h = sha256.New()
		h.Write(bsr)
		bsr = h.Sum(nil)
	}
	return bsr
}

// Test is an exported function to test the hash puzzle
func Test() {
	getHashPuzzles()
}

func getHashPuzzles() {
	// Define a seed value for testing
	seed := "I am a cryptographic hash puzzle"

	// Convert the seed to a byte slice
	seedBytes := []byte(seed)

	// Define a salt value for testing
	salt := []byte("123456")

	// Generate a hash puzzle based on the seed and salt
	hash := HashNb(seedBytes, 5, salt)

	// Print the result
	fmt.Printf("Hash: %x\n", hash)
}
