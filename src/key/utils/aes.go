package keyutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"log"
)

type AESUtils struct {
	aesPass   []byte
	block     cipher.Block
	encrypter cipher.BlockMode
	decrypter cipher.BlockMode
}

func NewAESUtils(blockSize int, mode string) *AESUtils {
	aesPass := []byte("test pass phrase")
	block, err := aes.NewCipher(aesPass)

	var encrypter cipher.BlockMode
	var decrypter cipher.BlockMode
	switch mode {
	case "CBC":
		iv := []byte("iv: to be random") //16
		encrypter = cipher.NewCBCEncrypter(block, iv)
		decrypter = cipher.NewCBCDecrypter(block, iv)
	case "ECB":
		encrypter = ECBModeEncrypter{&EcbMode{block}}
		decrypter = ECBModeDecrypter{&EcbMode{block}}
	default:
		log.Println("AES mode fallback to ECB")
	}

	if err != nil {
		panic(err)
	}

	return &AESUtils{
		aesPass:   aesPass,
		block:     block,
		encrypter: encrypter,
		decrypter: decrypter,
	}
}

func (a *AESUtils) GetKey() []byte {
	return a.aesPass
}

type EcbMode struct {
	Block cipher.Block
}

func (e EcbMode) BlockSize() int {
	return aes.BlockSize
}

type ECBModeEncrypter struct {
	*EcbMode
}

func (e ECBModeEncrypter) CryptBlocks(dst, src []byte) {
	tmp := encryptECB(src, e.Block, e.BlockSize())
	copy(dst, tmp)
}

func NewDecryptor(blk cipher.Block, mode string) cipher.BlockMode {
	switch mode {
	case "CBC":
		return cipher.NewCBCDecrypter(blk, []byte("iv: to be random"))
	case "ECB":
		return ECBModeDecrypter{&EcbMode{Block: blk}}
	default:
		log.Println("AES mode fallback to ECB")
	}
	return ECBModeDecrypter{&EcbMode{Block: blk}}
}

type ECBModeDecrypter struct {
	*EcbMode
}

func (e ECBModeDecrypter) CryptBlocks(dst, src []byte) {
	tmp := decryptECB(src, e.Block, e.BlockSize())
	copy(dst, tmp)
}

func encryptECB(data []byte, block cipher.Block, blockSize int) []byte {
	i := 0
	dst := make([]byte, len(data)+(blockSize-len(data)%blockSize))
	for ; i+blockSize < len(data); i += blockSize {
		block.Encrypt(dst[i:i+blockSize], data[i:i+blockSize])
	}
	if len(data)%blockSize == 0 {
		block.Encrypt(dst[i:i+blockSize], data[i:i+blockSize])
	} else {
		block.Encrypt(dst[i:], paddingByBlockSize(data[i:], blockSize))
	}
	return dst
}

func (a *AESUtils) Encrypt(data []byte) []byte {
	//padding here
	data = paddingByBlockSize(data, aes.BlockSize)
	dst := make([]byte, len(data))
	a.encrypter.CryptBlocks(dst, data)
	return dst[:len(dst)]
}

func decryptECB(data []byte, blk cipher.Block, blkSize int) []byte {
	i := 0
	dst := make([]byte, len(data))
	for ; i+blkSize < len(data); i += blkSize {
		blk.Decrypt(dst[i:i+blkSize], data[i:i+blkSize])
	}
	return bytes.TrimRight(dst, "\x00")
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
