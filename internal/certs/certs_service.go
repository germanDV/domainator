package certs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/notifier"
	"github.com/germandv/domainator/internal/tlser"
)

type Service interface {
	Save(ctx context.Context, req RegisterReq) (Cert, error)
	GetAll(ctx context.Context, req GetAllReq) ([]Cert, error)
	Delete(ctx context.Context, req DeleteReq) error
	Update(ctx context.Context, req UpdateReq) (Cert, error)
	ProcessBatch(ctx context.Context, size int, ch chan<- notifier.Notification, logger *slog.Logger) error
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

	data := s.tlsClient.GetCertData(cert.Domain)
	now := time.Now().UTC()

	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		err := s.repo.UpdateWithError(context.Background(), req.UserID, req.ID, string(data.Status), now)
		if err != nil {
			return Cert{}, err
		}

		cert.UpdatedAt = now
		cert.Error = string(data.Status)
		c, err := repoToServiceAdapter(cert)
		if err != nil {
			return Cert{}, err
		}
		return c, nil
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return Cert{}, ErrInvalidIssuer
	}

	err = s.repo.Update(ctx, req.UserID, req.ID, data.Expiry, issuer.value, now)
	if err != nil {
		return Cert{}, err
	}

	cert.UpdatedAt = now
	cert.Issuer = issuer.value
	cert.ExpiresAt = data.Expiry
	cert.Error = ""

	c, err := repoToServiceAdapter(cert)
	if err != nil {
		return Cert{}, err
	}
	return c, nil
}

func (s *CertsService) ProcessBatch(
	ctx context.Context,
	size int,
	ch chan<- notifier.Notification,
	logger *slog.Logger,
) error {
	var certs []repoCert
	var err error
	var wg sync.WaitGroup

	first := true
	lastID := ""

	for first || len(certs) > 0 {
		first = false

		certs, err = s.repo.GetBatch(ctx, size, lastID)
		if err != nil {
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
				s.updateAndCheckExp(cert, ch, logger)
			}(cert)
		}
		wg.Wait()
	}

	return nil
}

func (s *CertsService) updateAndCheckExp(cert repoCert, ch chan<- notifier.Notification, logger *slog.Logger) {
	logger.Debug("checking cert", "id", cert.ID, "domain", cert.Domain)
	data := s.tlsClient.GetCertData(cert.Domain)
	now := time.Now().UTC()

	userID, err := common.ParseID(cert.UserID)
	if err != nil {
		logger.Debug("failed to parse user ID", "id", cert.ID, "userID", cert.UserID, "error", err.Error())
		return
	}

	certID, err := common.ParseID(cert.ID)
	if err != nil {
		logger.Debug("failed to parse ID", "id", cert.ID, "error", err.Error())
		return
	}

	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		err := s.repo.UpdateWithError(context.Background(), userID, certID, string(data.Status), now)
		if err != nil {
			logger.Debug("failed to UpdateWithError", "id", cert.ID, "status", string(data.Status), "error", err.Error())
			return
		}
		ch <- notifier.Notification{
			ID:     cert.ID,
			UserID: cert.UserID,
			Domain: cert.Domain,
			Status: string(data.Status),
			Hours:  0,
		}
		return
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return
	}

	err = s.repo.Update(context.Background(), userID, certID, data.Expiry, issuer.value, now)
	if err != nil {
		logger.Debug("failed to update cert", "id", cert.ID, "error", err.Error())
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
