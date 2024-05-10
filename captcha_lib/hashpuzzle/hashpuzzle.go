package hashpuzzle

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"math/rand"
	"strings"
)

var difficulty string = "000"

// Generate a nonce that results in a hash starting with "000"
func generateNonce(seed string) string {
	nonce := 0
	for {
		// Generate a candidate solution
		x := seed + fmt.Sprint(nonce)
		hx := hashX(x)

		// Check if the hash starts with "000"
		if strings.HasPrefix(hx, difficulty) {
			break
		}

		// If not, increment the nonce and try again
		nonce++
	}
	return fmt.Sprint(nonce)
}

// Generate a random string of a given length
func generateString(seed string, length int) string {
	s := HashNb([]byte(seed), 10, []byte("asdasd"))

	rand.Seed(s)

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String() // prepend the seed
}

// same as HashNs but for byte strings
// requires N >= 0
func HashNb(bs []byte, N uint16, salt []byte) int64 {
	var h hash.Hash
	for i := uint16(0); i < N; i++ {
		h = sha256.New()
		h.Write(bs)
		h.Write(salt)
		bs = h.Sum(nil)
	}
	return int64(binary.BigEndian.Uint64(bs[:8]))
}

func hashX(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Generate puzzle key
func GenerateHashKey(seed string) string {
	puzzle := generateString(seed, 10)
	fmt.Print("The Puzzle is :", puzzle, "\n\nEnter a nonce value which when appended makes the hash of the format \"000..\" : ")
	var input int
	fmt.Scanln(&input)

	user_solution := puzzle + fmt.Sprint(input)
	hx := hashX(user_solution)

	// Check if the hash starts with "000"
	if strings.HasPrefix(hx, difficulty) {
		fmt.Println("\n\n\tSolution accepted")
		nonce := generateNonce(puzzle)
		result := puzzle + fmt.Sprint(nonce)
		return result
	} else {
		fmt.Println("\n\n\tSolution not accepted, try again!")
	}
	return ""
}
