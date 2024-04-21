package db

import (
	"context"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

const VersionTable = "schema_version"

type DbMigrator struct {
	migrator *migrate.Migrator
}

func NewDbMigrator(connStr string, fsys fs.FS) (*DbMigrator, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	migrator, err := migrate.NewMigrator(ctx, conn, VersionTable)
	if err != nil {
		return nil, err
	}

	err = migrator.LoadMigrations(fsys)
	if err != nil {
		return nil, err
	}

	return &DbMigrator{migrator}, nil
}

func (m *DbMigrator) Status(ctx context.Context) (int32, error) {
	return m.migrator.GetCurrentVersion(ctx)
}

func (m *DbMigrator) Up(ctx context.Context) error {
	return m.migrator.Migrate(ctx)
}

func (m *DbMigrator) Down(ctx context.Context) error {
	curr, err := m.Status(ctx)
	if err != nil {
		return err
	}
	return m.migrator.MigrateTo(ctx, curr-1)
}
