package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrator, cleanup, err := setupMigrator()
		if err != nil {
			return err
		}
		defer cleanup()

		ctx := context.Background()
		applied, pending, err := migrator.Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "APPLIED MIGRATIONS")
		fmt.Fprintln(w, "ID\tAPPLIED AT")
		for _, m := range applied {
			fmt.Fprintf(w, "%s\t%s\n", m.ID, m.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "PENDING MIGRATIONS")
		for _, p := range pending {
			fmt.Fprintln(w, p)
		}
		w.Flush()

		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&dbURL, "db-url", "", "Database connection URL (postgres://user:pass@host:port/dbname)")
	rootCmd.AddCommand(statusCmd)
}
