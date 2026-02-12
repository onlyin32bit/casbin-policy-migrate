package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/casbin/casbin/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxadapter "github.com/onlyin32bit/casbin-pgx-pgxpool-adapter"
	casbinMigrate "github.com/onlyin32bit/casbin-policy-migrate"
	"github.com/onlyin32bit/casbin-policy-migrate/adapter/pgx"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/casbin_test?sslmode=disable"
		log.Printf("DATABASE_URL not set, using default: %s", dbURL)
	}

	// Connect to database
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	a, _ := pgxadapter.NewAdapterByDB(pool)
	e, _ := casbin.NewEnforcer("./casbin/rbac_model.conf", a)
	e.LoadPolicy()
	// Create adapter
	adapter := pgx.NewAdapter(pool)

	// Create migrator
	migrator := casbinMigrate.NewMigrator(adapter, "casbin/policy_migrations")

	ctx := context.Background()

	// Run migrations
	fmt.Println("Running migrations...")
	if err := migrator.Up(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Check status
	fmt.Println("\nMigration status:")
	applied, pending, err := migrator.Status(ctx)
	if err != nil {
		log.Fatalf("Failed to get status: %v", err)
	}

	fmt.Printf("Applied migrations: %d\n", len(applied))
	for _, m := range applied {
		fmt.Printf("  - %s (applied at %s)\n", m.ID, m.AppliedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("\nPending migrations: %d\n", len(pending))
	for _, p := range pending {
		fmt.Printf("  - %s\n", p)
	}

	fmt.Println("\nMigrations completed successfully!")
}
