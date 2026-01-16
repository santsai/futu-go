package cipher

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type RSA struct {
	privateKey *rsa.PrivateKey
}

func NewRSA(keyPEM []byte) (*RSA, error) {

	// decode pem
	block, _ := pem.Decode(keyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	// Parse the PKCS#1 private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &RSA{privateKey: privateKey}, nil
}

func (c *RSA) Encrypt(data []byte) ([]byte, error) {

	var cdata []byte
	pubKey := &c.privateKey.PublicKey
	blksz := pubKey.Size() - 11

	for start := 0; start < len(data); start += blksz {
		block := data[start:min(start+blksz, len(data))]
		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, block)
		if err != nil {
			return nil, err
		}
		cdata = append(cdata, encrypted...)
	}

	return cdata, nil
}

func (c *RSA) Decrypt(data []byte) ([]byte, error) {

	var pdata []byte
	blksz := c.privateKey.Size()

	for start := 0; start < len(data); start += blksz {
		block := data[start:min(start+blksz, len(data))]
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, block)
		if err != nil {
			return nil, err
		}
		pdata = append(pdata, decrypted...)
	}

	return pdata, nil
}
