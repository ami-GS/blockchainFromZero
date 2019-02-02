package key

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
)

type KeyManager struct {
	PrivateKey rsa.PrivateKey
}

func New() *KeyManager {
	log.Println("generate new key")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return &KeyManager{
		PrivateKey: *priv,
	}
}

func (k *KeyManager) GetAddress() string {
	return k.PrivateKey.PublicKey.N.String()
}

func (k *KeyManager) Sign(data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, &(k.PrivateKey), crypto.SHA256, hashed[:])
}

// TODO: key/utils?
func ExportRsaPrivateKeyAsPem(privkey *rsa.PrivateKey) []byte {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privkey)
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)
}

func ParseRsaPrivateKeyFromPem(privPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}
