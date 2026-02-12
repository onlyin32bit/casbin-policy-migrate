package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dir string

var rootCmd = &cobra.Command{
	Use:   "casbin-migrate",
	Short: "A CLI for managing Casbin policy migrations",
	Long:  `A CLI tool that helps you manage your Casbin policy migrations, including creating new migrations, applying them, and rolling back.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&dir, "migrations-dir", "casbin/policy_migrations", "Directory for Casbin policy migrations")
}
