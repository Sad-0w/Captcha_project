package main

import (
	"captcha/captcha_lib/chess"
	"captcha/captcha_lib/hashpuzzle"
	"captcha/captcha_lib/sudoku"
	zipenc "captcha/captcha_lib/zipenc"
	"flag"
	"fmt"
	"log"
	"os"
	// "fmt"
)

func main() {
	debugLib := flag.String("debug", "hashpuzzle", "the target library to debug (runs Test)")
	keystr := flag.String("key", "thisisthedefault", "the key to use for encryption or decryption")
	N := flag.Int("hashes", 1000, "the number of hashes to perform on the key string")
	decorenc := flag.Bool("enc", true, "encrypt (true), or decrypt (false)")
	target := flag.String("in", "hhgttg.txt", "the file to zip and encrypt or decrypt and unzip")
	dest := flag.String("out", "hhgttg.bin", "the destination file or folder")
	// for testing if decryption is enabled the destination must be a folder (based on the unzip method used)

	flag.Parse()

	var PuzzleKey [3]string
	var K string
	var offsets []int
	offsets = nil
	PuzzleKey[0] = ""
	PuzzleKey[1] = ""
	PuzzleKey[2] = ""
	switch *debugLib {
	case "sudoku":
		PuzzleKey[0] = sudoku.GetPuzzleKey(*keystr, uint16(*N))
		fmt.Println(K)
		// return
	case "chess":
		PuzzleKey[1], offsets = chess.GetPuzzleKey(*keystr, nil)
		return
	case "hashpuzzle":
		PuzzleKey[2] = hashpuzzle.GenerateHashKey(*keystr)
		if len(PuzzleKey) > 0 {
			fmt.Println(PuzzleKey[2])
		}
		return
	}

	if *decorenc {
		err := zipenc.ZipAndEncrypt(&K, PuzzleKey, uint16(*N), *target, *dest, offsets)
		if err != nil {
			//Print error message:
			log.Println(err)
			os.Exit(-2)
		}
	} else {
		err := zipenc.DecryptAndUnzip(&K, *target, *dest)
		if err != nil {
			//Print error message:
			log.Println(err)
			os.Exit(-2)
		}
	}

}
