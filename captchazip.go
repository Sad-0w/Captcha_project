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
	decorenc := flag.Bool("enc", true, "encrypt (true), or decrypt (false)")
	target := flag.String("in", "hhgttg.txt", "the file to zip and encrypt or decrypt and unzip")
	dest := flag.String("out", "hhgttg.bin", "the destination file")

	flag.Parse()


	chess.Test()
	graphcolor.Test()

	key := zipenc.HashNs(*keystr, *N)
	if (*decorenc){
		err := zipenc.ZipAndEncrypt(key, *target, *dest)
	} else {
		err := zipenc.DecryptAndUnzip(key, *target, *dest)
	}
	if err != nil {
	 	//Print error message:
		log.Println(err)
		os.Exit(-2)
	}
}
