package main

import (
	chess "captcha/captcha_lib/chess"
	graphcolor "captcha/captcha_lib/graphcolor"
	zipenc "captcha/captcha_lib/zipenc"
	"flag"
	"log"
	"os"
	// "fmt"
)

func main() {
	keystr := flag.String("key", "thisisthedefault", "the key to use for encryption or decryption")
	N := flag.Int("hashes", 1000, "the number of hashes to perform on the key string")
	decorenc := flag.Bool("enc", true, "encrypt (true), or decrypt (false)")
	target := flag.String("in", "hhgttg.txt", "the file to zip and encrypt or decrypt and unzip")
	dest := flag.String("out", "hhgttg.bin", "the destination file or folder")
	// for testing if decryption is enabled the destination must be a folder (based on the unzip method used)

	flag.Parse()

	chess.Test()
	graphcolor.Test()

	if *decorenc {
		err := zipenc.ZipAndEncrypt(keystr, uint16(*N), *target, *dest)
		if err != nil {
			//Print error message:
			log.Println(err)
			os.Exit(-2)
		}
	} else {
		err := zipenc.DecryptAndUnzip(keystr, *target, *dest)
		if err != nil {
			//Print error message:
			log.Println(err)
			os.Exit(-2)
		}
	}

}
