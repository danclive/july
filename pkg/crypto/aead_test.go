package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	aead, err := NewAead(ChaCha20Poly1305, "abc")
	if err != nil {
		panic(err)
	}

	in := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	enc, err := aead.Encrypt(in)
	if err != nil {
		panic(err)
	}

	dec, err := aead.Decrypt(enc)
	if err != nil {
		panic(err)
	}

	assert.Exactly(t, in, dec)
}
