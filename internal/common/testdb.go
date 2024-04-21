package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/germandv/domainator/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDB struct {
	once      sync.Once
	dbName    string
	pool      *pgxpool.Pool
	terminate func(context.Context) error
}

// TestDB is a single instance to be shared accross tests, to initialize and run the test container only once.
var TestDB = &testDB{
	once:   sync.Once{},
	dbName: "domainator",
}

func (tdb *testDB) init() {
	fmt.Println("Running Postgres Test Container")

	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase(tdb.dbName),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		panic(fmt.Errorf("failed to strat posgres container: %w", err))
	}

	conn, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get postgres connection string: %w", err))
	}

	fmt.Printf("Postgres Test Container ConnectionString: %s\n", conn)
	dbPool, err := db.InitWithConnStr(conn)
	if err != nil {
		panic(fmt.Errorf("failed to init postgres: %w", err))
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to ping postgres: %w", err))
	}

	migrator, err := db.NewDbMigrator(conn, os.DirFS(filepath.Join("..", "..", "migrations")))
	if err != nil {
		panic(fmt.Errorf("failed create DB Migrator: %w", err))
	}

	err = migrator.Up(ctx)
	if err != nil {
		panic(fmt.Errorf("failed run migrations: %w", err))
	}

	tdb.pool = dbPool
	tdb.terminate = pgContainer.Terminate
}

// GetPool returns a Postgres connection pool.
// It starts the test container if it hasn't been started yet.
// It runs the migrations.
func (tdb *testDB) GetPool() *pgxpool.Pool {
	tdb.once.Do(tdb.init)
	return tdb.pool
}

func (tdb *testDB) StopContainer() error {
	return tdb.terminate(context.Background())
}
