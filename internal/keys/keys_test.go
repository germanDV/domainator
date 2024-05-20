package keys

import (
	"testing"
)

func TestKeypair(t *testing.T) {
	t.Parallel()

	var pemPriv string
	var pemPubl string

	t.Run("generate_key-pair", func(t *testing.T) {
		var err error
		pemPriv, pemPubl, err = NewPair()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("decode_keys", func(t *testing.T) {
		privKey, err := DecodePrivate(pemPriv)
		if err != nil {
			t.Error(err)
		}

		publKey, err := DecodePublic(pemPubl)
		if err != nil {
			t.Error(err)
		}

		if !privKey.PublicKey.Equal(publKey) {
			t.Error("keys do not match")
		}
	})
}
