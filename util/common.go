package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
	"unicode"

	"github.com/juju/errors"
)

// AesEncrypt aes 加密.
func AesEncrypt(encodeStr []byte, key []byte) (string, error) {
	encodeBytes := encodeStr
	block, err := aes.NewCipher(key)
	if err != err {
		return "", errors.Trace(err)
	}
	blockSize := block.BlockSize()
	encodeBytes = pkcs5Padding(encodeBytes, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(encodeBytes))
	blockMode.CryptBlocks(crypted, encodeBytes)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

// AesDecrypt aes 解密.
func AesDecrypt(decodeStr string, key []byte) ([]byte, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(decodeStr)
	if err != nil {
		return nil, errors.Trace(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Trace(err)
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(decodeBytes))
	blockMode.CryptBlocks(origData, decodeBytes)
	origData, err = pkcs5UnPadding(origData)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return origData, nil
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

// TrimSplit 按sep拆分，并去掉空字符.
func TrimSplit(raw, sep string) []string {
	var ss []string
	for _, val := range strings.Split(raw, sep) {
		if s := strings.TrimSpace(val); s != "" {
			ss = append(ss, s)
		}
	}
	return ss
}

//FieldEscape 转换为小写下划线分隔
func FieldEscape(k string) string {
	buf := []byte{}
	up := true
	for _, c := range k {
		if unicode.IsUpper(c) {
			if !up {
				buf = append(buf, '_')
			}
			c += 32
			up = true
		} else {
			up = false
		}

		buf = append(buf, byte(c))
	}
	return string(buf)
}
