package key

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/ami-GS/blockchainFromZero/textbook/06/key/utils"
)

type KeyManager struct {
	PrivateKey rsa.PrivateKey
}

func NewWithKey(privKeyByte []byte) (*KeyManager, error) {
	key, err := keyutils.ParseRsaPrivateKeyFromPem(privKeyByte)
	if err != nil {
		return nil, err
	}
	return &KeyManager{
		PrivateKey: *key,
	}, nil

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

func (k *KeyManager) GetAddressByHexString() string {
	return hex.EncodeToString(k.PrivateKey.PublicKey.N.Bytes())
}

func (k *KeyManager) Sign(data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, &(k.PrivateKey), crypto.SHA256, hashed[:])
}

func (k *KeyManager) Verify(data, sign []byte) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(&k.PrivateKey.PublicKey, crypto.SHA256, hashed[:], sign)
}
