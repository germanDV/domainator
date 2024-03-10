package tokenauth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/germandv/domainator/internal/keys"
	"github.com/golang-jwt/jwt/v5"
)

const TokenExpiration = 8 * time.Hour

type Service interface {
	Generate(userId string) (string, error)
	Validate(tokenString string) (jwt.MapClaims, error)
}

type TokenAuth struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func New(priv string, publ string) (*TokenAuth, error) {
	privateKey, err := keys.DecodePrivate(priv)
	if err != nil {
		return nil, err
	}

	publicKey, err := keys.DecodePublic(publ)
	if err != nil {
		return nil, err
	}

	return &TokenAuth{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (t *TokenAuth) Generate(userID string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(TokenExpiration).Unix(),
		"aud": "domainator",
		"iss": "domainator",
	})

	signed, err := token.SignedString(t.privateKey)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (t *TokenAuth) Validate(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
