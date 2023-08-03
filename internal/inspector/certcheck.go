package inspector

import (
	"fmt"
	"time"
)

// startCertsLoop starts a loop that checks TLS certificates.
// (it immediately performs a check and then sets the interval)
func (i Inspector) startCertsLoop() {
	ticker := time.NewTicker(i.checkCertInterval)
	defer ticker.Stop()

	i.checkCerts()
	for range ticker.C {
		i.checkCerts()
	}
}

// TODO: implement
func (i Inspector) checkCerts() {
	fmt.Println("Checking certificates...")
}
