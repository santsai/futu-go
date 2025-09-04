package futu_test

import (
	"testing"

	"github.com/santsai/futu-go"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	should := require.New(t)

	key := []byte("3FA037BF519D18D5")
	c, err := futu.NewAES(key, nil)
	should.NoError(err)

	data := []byte("hello, world")
	ciphertext, _ := c.Encrypt(data)
	plaintext, _ := c.Decrypt(ciphertext)
	should.Equal(data, plaintext)

	_, err = futu.NewAES(nil, nil)
	should.Error(err)

	should.Nil(c.Decrypt(nil))
	ciphertext[len(ciphertext)-1] = 15
	should.Nil(c.Decrypt(ciphertext))
}
