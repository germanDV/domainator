package main

import (
	"fmt"

	"github.com/germandv/domainator/internal/common"
)

// Generates a 32-byte secret (useful for signing cookies)
func main() {
	fmt.Println(common.GenerateRandomString(32))
}
