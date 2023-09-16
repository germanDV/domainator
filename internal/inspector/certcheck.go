package inspector

import (
	"context"
	"crypto/tls"
	"domainator/internal/certs"
	"domainator/internal/certstatus"
	"domainator/internal/config"
	"domainator/internal/logger"
	"domainator/internal/taskpool"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// doCertChecks gets all domains from the database and checks their certificates.
func (i Inspector) doCertChecks(doneCh chan<- struct{}) {
	domains, err := i.certsRepo.GetAll(context.Background())
	if err != nil {
		logger.Writer.Error(err)
		return
	}

	domainCount := len(domains)
	logger.Writer.Info("Domains to check: ", domainCount)

	workerCount := min(domainCount, 50)
	pool := taskpool.New(workerCount)

	for _, d := range domains {
		dd := d
		pool.Add(func() { i.checkCert(dd) })
	}

	pool.Wait()
	doneCh <- struct{}{}
}

// checkCert checks the certificate of a domain.
func (i Inspector) checkCert(c *certs.Cert) {
	logger.Writer.Info(fmt.Sprintf("Checking cert for %q", c.Domain))

	conn, err := tls.DialWithDialer(i.dialer, "tcp", c.Domain+":443", nil)
	if err != nil {
		i.saveAndSendBadCert(c, certstatus.CannotConnect, time.Time{})
		return
	}
	defer conn.Close()

	now := time.Now()
	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter

	err = conn.VerifyHostname(c.Domain)
	if err != nil {
		i.saveAndSendBadCert(c, certstatus.CannotConnect, expiry)
		return
	}

	if expiry.Before(now) {
		i.saveAndSendBadCert(c, certstatus.Expired, expiry)
		return
	}

	threshold := now.Add(config.GetDuration("CERT_EXPIRY_THRESHOLD"))
	if expiry.Before(threshold) {
		i.saveAndSendBadCert(c, certstatus.AboutToExpire, expiry)
		return
	}

	i.saveCheck(c, certstatus.OK, expiry)
}

// saveAndSendBadCert saves the failed check to the db and send the bad cert to the channel.
func (i Inspector) saveAndSendBadCert(c *certs.Cert, status certstatus.Status, expiry time.Time) {
	err := i.saveCheck(c, status, expiry)
	if err != nil {
		logger.Writer.Error(fmt.Sprintf("Error saving cert check: %s", err))
		return
	}

	badCert := BadCert{
		CertID: c.ID,
		Domain: c.Domain,
		Expiry: expiry,
		Status: status,
		Time:   time.Now(),
	}
	i.badCertsCh <- badCert
}

// saveCheck saves a cert check to the database.
func (i Inspector) saveCheck(c *certs.Cert, status certstatus.Status, expiry time.Time) error {
	check := certs.Check{
		ID:         uuid.New(),
		CertID:     c.ID,
		RespStatus: status,
		Expiry:     expiry,
	}
	return i.certsRepo.SaveCheck(context.Background(), &check)
}
