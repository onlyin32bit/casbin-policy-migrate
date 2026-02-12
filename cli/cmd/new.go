package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var sequential bool

var newCmd = &cobra.Command{
	Use:   "new [migration_name]",
	Short: "Create a new migration file",
	Long:  `Creates a new CSV migration file with either a timestamp prefix (default) or sequential numbering.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		name = strings.ReplaceAll(name, " ", "_")
		name = strings.ReplaceAll(name, "-", "_")

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("migrations directory not exists: %w", err)
		}

		var filename string
		if sequential {
			num, err := getNextSequentialNumber(dir)
			if err != nil {
				return err
			}
			filename = fmt.Sprintf("%04d_%s.csv", num, name)
		} else {
			timestamp := time.Now().Format("20060102150405")
			filename = fmt.Sprintf("%s_%s.csv", timestamp, name)
		}

		filePath := filepath.Join(dir, filename)

		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create migration file: %w", err)
		}
		defer f.Close()

		fmt.Printf("Created new migration file: %s\n", filePath)
		return nil
	},
}

func getNextSequentialNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	maxNum := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		parts := strings.Split(filename, "_")
		if len(parts) > 0 {
			if num, err := strconv.Atoi(parts[0]); err == nil {
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}

	return maxNum + 1, nil
}

func init() {
	newCmd.Flags().BoolVar(&sequential, "sequential", false, "Use sequential numbering (0001, 0002, etc) instead of timestamp prefix")
	rootCmd.AddCommand(newCmd)
}
