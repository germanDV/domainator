package certs

import (
	"context"
	"time"

	"github.com/germandv/domainator/internal/tlser"
	"github.com/germandv/domainator/internal/workerpool"
	"github.com/jackc/pgx/v5"
)

type Task struct {
	tlsClient tlser.Client
	cert      Cert
	repo      Repo
	tx        pgx.Tx
}

func (t Task) Execute() {
	data := t.tlsClient.GetCertData(t.cert.Domain.value)

	e := ""
	if data.Status != tlser.StatusOK && data.Status != tlser.StatusExpired {
		e = string(data.Status)
	}

	issuer, err := ParseIssuer(data.Issuer)
	if err != nil {
		return
	}

	now := time.Now().UTC()
	err = t.repo.UpdateWithTx(context.Background(), t.tx, t.cert.UserID, t.cert.ID, data.Expiry, issuer.value, now, e)
	if err != nil {
		return
	}
}

type Batch struct {
	tx pgx.Tx
	wp *workerpool.WorkerPool[Task]
}

func NewBatch(tasks []Task, tx pgx.Tx, concurrency int) *Batch {
	return &Batch{
		tx: tx,
		wp: workerpool.New[Task](tasks, concurrency),
	}
}

func (b *Batch) Begin() {
	b.wp.Run()
	defer b.tx.Commit(context.Background())
}
