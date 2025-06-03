package suite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type TestDatabase struct {
	Container *postgres.PostgresContainer
	DB        *sqlx.DB
	DSN       string
}

type dbOptions struct {
	migrationsPath string
}

type Option func(*dbOptions)

func WithMigrations(path string) Option {
	return func(opt *dbOptions) {
		opt.migrationsPath = path
	}
}

func NewTestDB(t testing.TB, opts ...Option) *TestDatabase {
	t.Helper()
	return createDatabase(t, opts...)
}

func createDatabase(t testing.TB, opts ...Option) *TestDatabase {
	t.Helper()
	ctx := context.Background()

	cfg := applyOptions(opts...)

	container := startPostgresContainer(t, ctx)

	dsn := getConnectionString(t, ctx, container)

	db := connectToDB(t, dsn)

	waitForDBReady(t, db)

	runMigrations(t, db, cfg)

	return &TestDatabase{
		Container: container,
		DB:        db,
		DSN:       dsn,
	}
}

func (tdb *TestDatabase) Close(ctx context.Context) {
	if err := tdb.DB.Close(); err != nil {
		zap.L().Warn("failed to close db", zap.Error(err))
	}
	if err := tdb.Container.Terminate(ctx); err != nil {
		zap.L().Error("failed to terminate container", zap.Error(err))
	}
}

func waitForPing(db *sqlx.DB) error {
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("failed to connect to database after retries")
}

func applyOptions(opts ...Option) dbOptions {
	var cfg dbOptions
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func startPostgresContainer(t testing.TB, ctx context.Context) *postgres.PostgresContainer {
	container, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("db_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithPollInterval(1*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	return container
}

func getConnectionString(t testing.TB, ctx context.Context, container *postgres.PostgresContainer) string {
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")
	return dsn
}

func connectToDB(t testing.TB, dsn string) *sqlx.DB {
	db, err := sqlx.Connect("pgx", dsn)
	require.NoError(t, err, "failed to connect to db")
	return db
}

func waitForDBReady(t testing.TB, db *sqlx.DB) {
	require.NoError(t, waitForPing(db), "database did not become available")
}

func runMigrations(t testing.TB, db *sqlx.DB, cfg dbOptions) {
	if cfg.migrationsPath == "" {
		return
	}
	goose.SetDialect("postgres")
	err := goose.Up(db.DB, cfg.migrationsPath)
	require.NoError(t, err, "failed to apply migrations")
}
