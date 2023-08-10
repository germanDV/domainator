// Package main implements a simple CLI to perform secondary tasks.
package main

import (
	"domainator/internal/keys"
	"fmt"
)

func main() {
	// This could potentially hold several commands,
	// but for now it's just used to generate a key-pair
	// to be used for signing and verifying JWTs.
	// So we invoke newKeyPair() directly.
	newKeyPair()
}

func newKeyPair() {
	priv, publ, err := keys.NewPair()
	if err != nil {
		panic(err)
	}
	fmt.Println(priv)
	fmt.Println(publ)
}
