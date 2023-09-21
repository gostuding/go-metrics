package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

var (
	keySize              = 4096
	fileBits os.FileMode = 0644
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Fatalf("generate keys error: %v", err)
	}
	pubASN1, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		log.Fatalf("marshal public key error: %v", err)
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "AGENT PUBLIC KEY",
		Bytes: pubASN1,
	})
	if err = os.WriteFile("public_key", pubBytes, fileBits); err != nil {
		log.Fatalf("save public key error: %v", err)
	}
	prvBytes := x509.MarshalPKCS1PrivateKey(key)
	prvPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "SERVER PRIVATE KEY",
			Bytes: prvBytes,
		},
	)
	if err = os.WriteFile("private_key", prvPem, fileBits); err != nil {
		log.Fatalf("save private key error: %v", err)
	}
}
