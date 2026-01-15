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

	pubKey := &c.privateKey.PublicKey
	cdata, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		return nil, err
	}
	return cdata, nil
}

func (c *RSA) Decrypt(data []byte) ([]byte, error) {

	pdata, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, data)
	if err != nil {
		return nil, err
	}
	return pdata, nil
}
