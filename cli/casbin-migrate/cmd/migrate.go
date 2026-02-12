package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	casbinMigrate "github.com/onlyin32bit/casbin-policy-migrate"
	"github.com/onlyin32bit/casbin-policy-migrate/adapter/pgx"
	"github.com/spf13/cobra"
)

var dbURL string

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrator, cleanup, err := setupMigrator()
		if err != nil {
			return err
		}
		defer cleanup()

		if err := migrator.Up(context.Background()); err != nil {
			return err
		}
		fmt.Println("Migrations applied successfully!")
		return nil
	},
}

var downCmd = &cobra.Command{
	Use:   "down [n]",
	Short: "Rollback the last n migrations (default 1). Use 0 for all.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 1
		if len(args) > 0 {
			var err error
			_, err = fmt.Sscanf(args[0], "%d", &n)
			if err != nil {
				return fmt.Errorf("invalid number: %s", args[0])
			}
		}

		migrator, cleanup, err := setupMigrator()
		if err != nil {
			return err
		}
		defer cleanup()

		if err := migrator.Down(context.Background(), n); err != nil {
			return err
		}
		fmt.Println("Rollback successful!")
		return nil
	},
}

func setupMigrator() (*casbinMigrate.Migrator, func(), error) {
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		return nil, nil, fmt.Errorf("database URL is required (use --db-url or DATABASE_URL env var)")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cleanup := func() {
		pool.Close()
	}

	adapter := pgx.NewAdapter(pool)
	migrator := casbinMigrate.NewMigrator(adapter, dir)

	return migrator, cleanup, nil
}

func init() {
	upCmd.Flags().StringVar(&dbURL, "db-url", "", "Database connection URL (postgres://user:pass@host:port/dbname)")
	downCmd.Flags().StringVar(&dbURL, "db-url", "", "Database connection URL (postgres://user:pass@host:port/dbname)")
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}
