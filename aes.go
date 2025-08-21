package futu

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// CryptoAES is a struct for encrypting and decrypting data using AES CBC mode.
type CryptoAES struct {
	block cipher.Block
	iv    []byte
}

// NewCryptoAES creates a new CryptoAES instance.
func NewCryptoAES(key, iv []byte) (*CryptoAES, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if iv == nil {
		iv = make([]byte, aes.BlockSize)
	}

	return &CryptoAES{block: block, iv: iv}, nil
}

// Encrypt encrypts the data.
func (c *CryptoAES) Encrypt(data []byte) []byte {
	data = addPKCS7Padding(data)
	ciphertext := make([]byte, len(data))
	cipher.NewCBCEncrypter(c.block, c.iv).CryptBlocks(ciphertext, data)

	return ciphertext
}

// Decrypt decrypts the data.
func (c *CryptoAES) Decrypt(data []byte) []byte {
	plaintext := make([]byte, len(data))
	cipher.NewCBCDecrypter(c.block, c.iv).CryptBlocks(plaintext, data)

	return removePKCS7Padding(plaintext)
}

func addPKCS7Padding(data []byte) []byte {
	paddingLen := aes.BlockSize - len(data)%aes.BlockSize
	padding := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)

	return append(data, padding...)
}

func removePKCS7Padding(data []byte) []byte {
	length := len(data)

	if length == 0 {
		return nil
	}
	paddingLen := int(data[length-1])
	if paddingLen > length {
		return nil
	}

	return data[:length-paddingLen]
}
