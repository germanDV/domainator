package certs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/notifier"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	Save(ctx context.Context, req RegisterReq) (Cert, error)
	GetAll(ctx context.Context, req GetAllReq) ([]Cert, error)
	Delete(ctx context.Context, req DeleteReq) error
	Update(ctx context.Context, req UpdateReq) (Cert, error)
	ProcessBatch(ctx context.Context, size int, concurrency int, ch chan<- notifier.Notification) error
}

type CertsService struct {
	repo            Repo
	tlsClient       tlser.Client
	maxCertsPerUser int
}

func NewService(tlsClient tlser.Client, repo Repo, maxCertsPerUser int) *CertsService {
	return &CertsService{
		repo:            repo,
		tlsClient:       tlsClient,
		maxCertsPerUser: maxCertsPerUser,
	}
}

func (s *CertsService) Save(ctx context.Context, req RegisterReq) (Cert, error) {
	count, err := s.repo.Count(ctx, req.UserID, s.maxCertsPerUser)
	if err != nil {
		return Cert{}, err
	}

	if count >= s.maxCertsPerUser {
		return Cert{}, fmt.Errorf("cannot have more than %d certs", s.maxCertsPerUser)
	}

	data := s.tlsClient.GetCertData(req.Domain.value)
	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		return Cert{}, fmt.Errorf("TLS error: %s", data.Status)
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return Cert{}, err
	}

	cert := New(req.UserID, req.Domain, issuer, data.Expiry)
	err = s.repo.Save(ctx, serviceToRepoAdapter(cert))
	if err != nil {
		return Cert{}, err
	}

	return cert, nil
}

func (s *CertsService) GetAll(ctx context.Context, req GetAllReq) ([]Cert, error) {
	certificates, err := s.repo.GetAll(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	certs := make([]Cert, len(certificates))
	for i, c := range certificates {
		cert, err := repoToServiceAdapter(c)
		if err != nil {
			return nil, err
		}
		certs[i] = cert
	}

	return certs, nil
}

func (s *CertsService) Delete(ctx context.Context, req DeleteReq) error {
	return s.repo.Delete(ctx, req.UserID, req.ID)
}

func (s *CertsService) Update(ctx context.Context, req UpdateReq) (Cert, error) {
	cert, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return Cert{}, err
	}

	e := ""
	data := s.tlsClient.GetCertData(cert.Domain)
	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		e = string(data.Status)
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return Cert{}, ErrInvalidIssuer
	}

	now := time.Now().UTC()
	err = s.repo.Update(ctx, req.UserID, req.ID, data.Expiry, issuer.value, now, e)
	if err != nil {
		return Cert{}, err
	}

	cert.UpdatedAt = now
	cert.Issuer = issuer.value
	cert.ExpiresAt = data.Expiry
	cert.Error = e

	c, err := repoToServiceAdapter(cert)
	if err != nil {
		return Cert{}, err
	}
	return c, nil
}

func (s *CertsService) ProcessBatch(
	ctx context.Context,
	size int,
	concurrency int,
	ch chan<- notifier.Notification,
) error {
	var certs []repoCert
	var err error
	var wg sync.WaitGroup

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}

	first := true
	lastID := ""

	for first || len(certs) > 0 {
		first = false

		certs, err = s.repo.ProcessBatch(ctx, tx, size, lastID)
		if err != nil {
			e := tx.Rollback(context.Background())
			if e != nil {
				return fmt.Errorf("an error occurred: %w. And tx did not rollback successfully: %w", err, e)
			}
			return err
		}
		if len(certs) == 0 {
			break
		}

		wg.Add(len(certs))
		for _, cert := range certs {
			lastID = cert.ID
			go func(cert repoCert) {
				defer wg.Done()
				s.updateAndCheckExp(cert, tx, ch)
			}(cert)
		}
		wg.Wait()
	}

	return tx.Commit(context.Background())
}

func (s *CertsService) updateAndCheckExp(cert repoCert, tx pgx.Tx, ch chan<- notifier.Notification) {
	data := s.tlsClient.GetCertData(cert.Domain)

	e := ""
	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		e = string(data.Status)
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return
	}

	userID, err := common.ParseID(cert.UserID)
	if err != nil {
		return
	}

	certID, err := common.ParseID(cert.ID)
	if err != nil {
		return
	}

	now := time.Now().UTC()

	err = s.repo.UpdateWithTx(context.Background(), tx, userID, certID, data.Expiry, issuer.value, now, e)
	if err != nil {
		return
	}

	expHours := hoursToExpiration(data.Expiry)
	expStatus := expirationStatus(expHours)
	if expStatus != "" {
		ch <- notifier.Notification{
			ID:     cert.ID,
			UserID: cert.UserID,
			Domain: cert.Domain,
			Status: expStatus,
			Hours:  expHours,
		}
	}
}

func hoursToExpiration(expiry time.Time) int {
	return int(expiry.Sub(time.Now().UTC()).Hours())
}

func expirationStatus(hours int) string {
	if hours <= 0 {
		return "expired"
	}
	if hours < 24 {
		return "expires today"
	}
	if hours < 72 {
		return "expires soon"
	}
	return ""
}
