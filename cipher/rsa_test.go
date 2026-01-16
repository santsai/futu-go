package cipher_test

import (
	"testing"

	"github.com/santsai/futu-go/cipher"
	"github.com/stretchr/testify/require"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func genPEM() []byte {
	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	der := x509.MarshalPKCS1PrivateKey(priv)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: der,
	})

	return pemBytes
}

func TestRSA(t *testing.T) {
	should := require.New(t)

	key := genPEM()

	c, err := cipher.NewRSA(key)
	should.NoError(err)

	data := key
	ciphertext, _ := c.Encrypt(data)
	plaintext, _ := c.Decrypt(ciphertext)
	should.Equal(data, plaintext)

	_, err = cipher.NewRSA(nil)
	should.Error(err)

	should.Nil(c.Decrypt(nil))
}
