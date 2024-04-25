package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log"
	"os"
)

var (
	public_key = GenerateKey()
)

func GenerateKey() *rsa.PublicKey {
	publicKey, _ := os.ReadFile("./rsa_public_key.pem")
	block, _ := pem.Decode([]byte(publicKey))
	if block.Type != "PUBLIC KEY" {
		log.Fatal("error decoding public key from pem")
	}
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal("error parsing key")
	}
	var ok bool
	var pubkey *rsa.PublicKey
	if pubkey, ok = parsedKey.(*rsa.PublicKey); !ok {
		log.Fatal("unable to parse public key")
	}

	return pubkey
}

func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) string {
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, pub, msg, nil)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)

}

func chunkSlice(slice []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func Encrypt(msg []byte) []string {
	res := []string{}
	for _,chunk  := range chunkSlice(msg, 400) {
		res = append(res, EncryptWithPublicKey(chunk, public_key))
	}
	return res
}
