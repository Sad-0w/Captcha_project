package zipenc

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ContextHeaderStruct struct {
	N    uint16
	Salt []byte
}

func zipFile(infile string, outfile string) error {
	// create zip file
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	// create zip writer
	writer := zip.NewWriter(f)
	defer writer.Close()

	// walk through subdirectories (if any)
	return filepath.Walk(infile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// create zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// choose compression option
		header.Method = zip.Deflate

		// grab the filename
		header.Name, err = filepath.Rel(filepath.Dir(infile), path)
		if err != nil {
			return err
		}

		// mark the directories with proper notation
		if info.IsDir() {
			header.Name += "/"
		}

		// create header writer for zip file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func unzipFile(infile string, outfile string) error {
	r, err := zip.OpenReader(infile)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(outfile, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(outfile, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(outfile)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// SHA256 Hash function
// input a string outputs the sha256 bytes
func Hashs(s string, salt []byte) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	h.Write(salt)
	bs := h.Sum(nil)
	return bs
}

// N times SHA256 Hash function
// input a string outputs the sha256 bytes
// H(H(H(H(...string))))
// requires N > 0
func HashNs(s string, N uint16, salt []byte) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	h.Write(salt)
	bs := h.Sum(nil)
	for i := uint16(1); i < N; i++ {
		h = sha256.New()
		h.Write(bs)
		bs = h.Sum(nil)
	}
	return bs
}

// same as Hashs but for byte strings
func Hashb(bs []byte, salt []byte) []byte {
	h := sha256.New()
	h.Write(bs)
	h.Write(salt)
	bsr := h.Sum(nil)
	return bsr
}

// same as HashNs but for byte strings
// requires N >= 0
func HashNb(bs []byte, N uint16, salt []byte) []byte {
	bsr := bs
	var h hash.Hash
	h = sha256.New()
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

// Encrypt a file with AES GCM mode
// takes in key, input filename, output filename
func encrypt(keystr *string, N uint16, infile string, outfile string) (err error) {

	salt := make([]byte, 16)
	rand.Read(salt)
	key := HashNs(*keystr, N, salt)

	// get the file plaintext
	plainText, err := os.ReadFile(infile)
	if err != nil {
		log.Fatalf("read file err: %v", err.Error())
		return err
	}

	// create a new AES cipher using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("cipher err: %v", err.Error())
		return err
	}

	// use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("cipher GCM err: %v", err.Error())
		return err
	}

	// make nonce with gcm mode
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		log.Fatalf("nonce  err: %v", err.Error())
		return err
	}

	// encrypt file
	cipherText := gcm.Seal(nonce, nonce, plainText, nil)

	// create the context header used to create the key
	header := make([]byte, 32)
	copy(header[:16], salt[:])
	binary.LittleEndian.PutUint16(header[16:], N)
	encrypted := append(header, cipherText...)
	// write file to output
	err = os.WriteFile(outfile, encrypted, 0777)
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
		return err
	}
	return nil
}

// parse the context header used to create the key
func parseHeader(text []byte) ContextHeaderStruct {
	var header ContextHeaderStruct

	header.Salt = text[:16]
	header.N = binary.LittleEndian.Uint16(text[16:])
	return header
}

// Decrypts a file with AES GCM mode
// takes in key, input filename, output filename
func decrypt(keystr string, infile string, outfile string) (err error) {

	// get the file plaintext
	cipherText, err := os.ReadFile(infile)
	if err != nil {
		log.Fatal(err)
		return err
	}

	header := parseHeader(cipherText[:32])
	key := HashNs(keystr, header.N, header.Salt)

	cipherText = cipherText[32:]

	// create AES block
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("cipher err: %v", err.Error())
		return err
	}

	// use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("cipher GCM err: %v", err.Error())
		return err
	}

	// remove nonce and decrypt
	nonce := cipherText[:gcm.NonceSize()]
	cipherText = cipherText[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		log.Fatalf("decrypt file err: %v", err.Error())
		return err
	}

	// write file to output
	err = os.WriteFile(outfile, plainText, 0777)
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
		return err
	}
	return nil
}

func ZipAndEncrypt(keystr *string, N uint16, infile string, outfile string) (err error) {

	err = zipFile(infile, infile+".zip")
	if err != nil {
		log.Fatalf("zip err: %v", err.Error())
		return err
	}
	err = encrypt(keystr, N, infile+".zip", outfile)
	if err != nil {
		log.Fatalf("encrypt err: %v", err.Error())
		return err
	}
	err = os.Remove(infile + ".zip")
	if err != nil {
		log.Fatalf("remove file err: %v", err.Error())
		return err
	}
	return nil
}

func DecryptAndUnzip(keystr *string, infile string, outfile string) (err error) {
	err = decrypt(*keystr, infile, infile+".zip")
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
		return err
	}
	err = unzipFile(infile+".zip", outfile)
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
		return err
	}
	err = os.Remove(infile + ".zip")
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
		return err
	}
	return nil
}
