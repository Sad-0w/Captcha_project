package main

import (
	zipenc "captcha/zipenc"
	"log"
	"os"
	// "fmt"
)

func main() {
	key := zipenc.HashNs("this is a test", 1000)
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
