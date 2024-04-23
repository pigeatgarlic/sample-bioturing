package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
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
