package keyutils

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
)

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
		return nil, errors.New("failed to parse private key PEM block")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ParseRsaPublicKeyFromPem(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse public key PEM block")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub, nil
}

func GenPublicKey(keySeq []byte) *rsa.PublicKey {
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(keySeq),
		E: 65573,
	}
}

func VerifySignRSA256(data, sign []byte, pubKey *rsa.PublicKey) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], sign)
}
