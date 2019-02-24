package keyutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

type AESUtils struct {
	blockSize int
	aesPass   []byte
	cipher    cipher.Block
}

func NewAESUtils(blockSize int) *AESUtils {
	aesPass := []byte("test pass phrase")
	block, err := aes.NewCipher(aesPass)
	if err != nil {
		panic(err)
	}

	return &AESUtils{
		blockSize: blockSize,
		aesPass:   aesPass,
		cipher:    block,
	}
}

func (a *AESUtils) GetKey() []byte {
	return a.aesPass
}

// not secure encrypt
func (a *AESUtils) EncryptSplit(data []byte) []byte {
	i := 0
	dst := make([]byte, len(data)+(a.blockSize-len(data)%a.blockSize))
	for ; i+a.blockSize < len(data); i += a.blockSize {
		a.cipher.Encrypt(dst[i:i+a.blockSize], data[i:i+a.blockSize])
	}
	if len(data)%a.blockSize == 0 {
		a.cipher.Encrypt(dst[i:i+a.blockSize], data[i:i+a.blockSize])
	} else {
		a.cipher.Encrypt(dst[i:], paddingByBlockSize(data[i:], a.blockSize))
	}
	return dst
}

func DecryptSplit(data []byte, blk cipher.Block, blkSize int) []byte {
	i := 0
	dst := make([]byte, len(data))
	for ; i+blkSize < len(data); i += blkSize {
		blk.Decrypt(dst[i:i+blkSize], data[i:i+blkSize])
	}
	return bytes.TrimRight(dst, "\x00")
}
func (a *AESUtils) DecryptSplit(data []byte) []byte {
	return DecryptSplit(data, a.cipher, a.blockSize)
}

func (a *AESUtils) DecryptSplitWithKey(data, aesPass []byte) []byte {
	block, _ := aes.NewCipher(aesPass)
	return DecryptSplit(data, block, a.blockSize)
}

func paddingByBlockSize(data []byte, blockSize int) []byte {
	return append(data, make([]byte, blockSize-len(data))...)
}

// TODO: can be more generic utility func
func EncodeBase64(data []byte) []byte {
	outLen := base64.StdEncoding.EncodedLen(len(data))
	dst := make([]byte, outLen)
	base64.StdEncoding.Encode(dst, data)
	return dst
}

// TODO: can be more generic utility func
func DecodeBase64(data []byte) ([]byte, error) {
	outLen := base64.StdEncoding.DecodedLen(len(data))
	dst := make([]byte, outLen)
	_, err := base64.StdEncoding.Decode(dst, data)
	return dst, err
}
