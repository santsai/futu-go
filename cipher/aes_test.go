package cipher_test

import (
	"testing"

	"github.com/santsai/futu-go/cipher"
	"github.com/stretchr/testify/require"
)

func TestAES(t *testing.T) {
	should := require.New(t)

	key := []byte("3FA037BF519D18D5")
	c, err := cipher.NewAES(key, nil)
	should.NoError(err)

	data := []byte("hello, world")
	ciphertext, _ := c.Encrypt(data)
	plaintext, _ := c.Decrypt(ciphertext)
	should.Equal(data, plaintext)

	_, err = cipher.NewAES(nil, nil)
	should.Error(err)

	should.Nil(c.Decrypt(nil))
	ciphertext[len(ciphertext)-1] = 15
	should.Nil(c.Decrypt(ciphertext))
}
