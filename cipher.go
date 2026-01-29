package futu

import (
	"github.com/santsai/futu-go/cipher"
	"github.com/santsai/futu-go/pb"
)

// cipherManager holds the RSA and AES cipher instances
type cipherManager struct {
	rsa *cipher.RSA
	aes *cipher.AES
}

// newCipherManager creates a new cipherManager
// If privateKey is nil, returns nil
func newCipherManager(privateKey []byte) (*cipherManager, error) {
	if privateKey == nil {
		return nil, nil
	}

	rsa, err := cipher.NewRSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &cipherManager{
		rsa: rsa,
	}, nil
}

// Encrypt encrypts the provided data using the appropriate cipher based on ProtoId
func (cm *cipherManager) Encrypt(protoId pb.ProtoId, data []byte) ([]byte, error) {

	if c := cm.getCipher(protoId); c != nil {
		return c.Encrypt(data)
	}

	return data, nil
}

// Decrypt decrypts the provided data using the appropriate cipher based on ProtoId
func (cm *cipherManager) Decrypt(protoId pb.ProtoId, data []byte) ([]byte, error) {

	if c := cm.getCipher(protoId); c != nil {
		return c.Decrypt(data)
	}

	return data, nil
}

// UpdateAES updates the AES parameters (key and IV)
func (cm *cipherManager) UpdateAES(key []byte, iv []byte) error {
	if cm == nil {
		return nil
	}

	aes, err := cipher.NewAES(key, iv)
	if err != nil {
		return err
	}

	cm.aes = aes
	return nil
}

// getCipher returns the appropriate cipher based on ProtoId
// Used internally by Encrypt and Decrypt methods
func (cm *cipherManager) getCipher(protoId pb.ProtoId) cipher.Cipher {
	if cm == nil {
		return nil
	}

	if protoId == pb.ProtoId_InitConnect {
		return cm.rsa
	}

	return cm.aes
}
