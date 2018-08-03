package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/juju/errors"
)

func byteKey(key string) []byte {
	for i := 16; i < 33; i += 8 {
		key += strings.Repeat("\x00", i-len(key))
		if len(key) == i {
			break
		}
	}
	return []byte(key)
}

// Encrypt aes 加密.
func Encrypt(encodeStr string, key string) (string, error) {
	encodeBytes := []byte(encodeStr)
	encodeKey := byteKey(key)
	block, err := aes.NewCipher(encodeKey)
	if err != nil {
		return "", errors.Trace(err)
	}
	blockSize := block.BlockSize()
	encodeBytes = pkcs5Padding(encodeBytes, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, encodeKey[:blockSize])
	crypted := make([]byte, len(encodeBytes))
	blockMode.CryptBlocks(crypted, encodeBytes)

	return base64.URLEncoding.EncodeToString(crypted), nil
}

// Decrypt aes 解密.
func Decrypt(decodeStr string, key string) (string, error) {
	decodeKey := byteKey(key)
	decodeBytes, err := base64.URLEncoding.DecodeString(decodeStr)
	if err != nil {
		return "", errors.Trace(err)
	}
	block, err := aes.NewCipher(decodeKey)
	if err != nil {
		return "", errors.Trace(err)
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, decodeKey[:blockSize])
	origData := make([]byte, len(decodeBytes))
	fmt.Printf("ori:%d, dec:%d\n", len(origData), len(decodeBytes))
	blockMode.CryptBlocks(origData, decodeBytes)
	origData, err = pkcs5UnPadding(origData)
	if err != nil {
		return "", errors.Trace(err)
	}

	return string(origData), nil
}

// pkcs5Padding pkcs5 添加数据.
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// pkcs5UnPadding pkcs5 删除数据.
func pkcs5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)

	if length <= 0 {
		return nil, errors.New("pkcs5Padding len(origData) <= 0 error")
	}

	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)], nil
}
