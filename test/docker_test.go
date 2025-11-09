package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pressly/goose/v3"
)

var (
	dockerPool *dockertest.Pool
	resource   *dockertest.Resource
	dbPool     *pgxpool.Pool
	ctx        = context.Background()
)

func TestMain(m *testing.M) {
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("No .env file found or error loading .env file: %v", err)
	}

	dockerPool, err = dockertest.NewPool(os.Getenv("DOCKER_HOST"))
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = dockerPool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	dockerPool.MaxWait = 120 * time.Second

	resource, err = dockerPool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	addr := resource.GetHostPort("5432/tcp")
	postgresDSN := fmt.Sprintf("postgres://postgres:password@%s/test?sslmode=disable", addr)
	log.Printf("Connecting to test database at %s", addr)

	if err = dockerPool.Retry(func() error {
		dbPool, err = pgxpool.New(ctx, postgresDSN)
		if err != nil {
			return err
		}
		return dbPool.Ping(ctx)
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	log.Println("Successfully connected to test database")

	if err = runMigrations(postgresDSN); err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}
	log.Println("Migrations completed successfully")

	code := m.Run()

	dbPool.Close()
	if err = dockerPool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func runMigrations(databaseUrl string) error {
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get the project root directory
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	migrationsDir := filepath.Join(projectRoot, "migrations", "postgres")

	if _, err = os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Printf("Migrations directory not found at %s, skipping migrations", migrationsDir)
		return nil
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("can not choose `postgres` as database dialect: %w", err)
	}

	if err = goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func GetPgxPool() *pgxpool.Pool {
	return dbPool
}

func GetPostgresDSN() string {
	addr := resource.GetHostPort("5432/tcp")
	return fmt.Sprintf("postgres://testuser:secret@%s/testdb?sslmode=disable", addr)
}
