// Package keys provides functionality around asymmetric keys.
package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

// NewPair generates a ecdsa key-pair and returns both the private and public keys in PEM format.
func NewPair() (string, string, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	publicKey := &privateKey.PublicKey

	privatePEM, err := encodePrivate(privateKey)
	if err != nil {
		return "", "", err
	}

	publicPEM, err := encodePublic(publicKey)
	if err != nil {
		return "", "", err
	}

	return privatePEM, publicPEM, nil
}

// encodePrivate encodes the private key to PEM format.
func encodePrivate(privKey *ecdsa.PrivateKey) (string, error) {
	encoded, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return "", err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded})
	key := string(pemEncoded)
	return key, nil
}

// encodePublic encondes the public key to PEM format.
func encodePublic(pubKey *ecdsa.PublicKey) (string, error) {
	encoded, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded})
	key := string(pemEncodedPub)
	return key, nil
}

// DecodePrivate decodes the private key from PEM format.
func DecodePrivate(pemEncodedPriv string) (*ecdsa.PrivateKey, error) {
	blockPriv, _ := pem.Decode([]byte(pemEncodedPriv))
	x509EncodedPriv := blockPriv.Bytes
	return x509.ParseECPrivateKey(x509EncodedPriv)
}

// DecodePublic decodes the public key from PEM format.
func DecodePublic(pemEncodedPub string) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey, nil
}
