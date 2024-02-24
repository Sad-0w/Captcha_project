package main

import (
	zipenc "captcha/captcha_lib/zipenc"
	chess "captcha/captcha_lib/chess"
	graphcolor "captcha/captcha_lib/graphcolor"
	"log"
	"os"
	"flag"
	// "fmt"
)

func main() {
	keystr := flag.String("key", "thisisthedefault", "the key to use for encryption or decryption")
	N := flag.Int("hashes", 1000, "the number of hashes to perform on the key string")

	flag.Parse()


	chess.Test()
	graphcolor.Test()

	key := zipenc.HashNs(*keystr, *N)
	err := zipenc.ZipAndEncrypt(key, "hhgttg.txt", "hhgttg.bin")
	if err != nil {
	 	//Print error message:
		log.Println(err)
		os.Exit(-2)
		}
	err = zipenc.DecryptAndUnzip(key, "hhgttg.bin", "hhgttg")
	if err != nil {
	 	//Print error message:
		log.Println(err)
		os.Exit(-2)
	}
}
