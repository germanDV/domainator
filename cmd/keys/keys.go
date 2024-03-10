package main

import (
	"fmt"

	"github.com/germandv/domainator/internal/keys"
)

// Generate key-pair
func main() {
	priv, publ, err := keys.NewPair()
	if err != nil {
		panic(err)
	}
	fmt.Println(priv)
	fmt.Println(publ)
}
