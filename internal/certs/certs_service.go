package certs

import (
	"context"
	"fmt"
	"time"

	"github.com/germandv/domainator/internal/notifier"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	Save(ctx context.Context, req RegisterReq) (Cert, error)
	GetAll(ctx context.Context, req GetAllReq) ([]Cert, error)
	Delete(ctx context.Context, req DeleteReq) error
	Update(ctx context.Context, req UpdateReq) (Cert, error)
	ProcessBatch(ctx context.Context, size int, concurrency int, notificationCh chan notifier.Notification) error
}

type CertsService struct {
	repo      Repo
	tlsClient tlser.Client
}

func NewService(tlsClient tlser.Client, repo Repo) *CertsService {
	return &CertsService{
		repo:      repo,
		tlsClient: tlsClient,
	}
}

func (s *CertsService) Save(ctx context.Context, req RegisterReq) (Cert, error) {
	data := s.tlsClient.GetCertData(req.Domain.value)
	if data.Status != tlser.StatusOK {
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
	notificationCh chan notifier.Notification,
) error {
	var certs []repoCert
	var tx pgx.Tx
	var err error

	first := true
	lastID := ""

	for first || len(certs) > 0 {
		first = false

		certs, tx, err = s.repo.ProcessBatch(ctx, size, lastID)
		if err != nil {
			return err
		}

		tasks := make([]Task, len(certs))
		for i, entry := range certs {
			cert, err := repoToServiceAdapter(entry)
			if err != nil {
				continue
			}

			tasks[i] = Task{
				cert:           cert,
				tlsClient:      s.tlsClient,
				repo:           s.repo,
				tx:             tx,
				notificationCh: notificationCh,
			}

			lastID = entry.ID
		}

		batch := NewBatch(tasks, tx, concurrency)
		batch.Begin()
	}

	return nil
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
