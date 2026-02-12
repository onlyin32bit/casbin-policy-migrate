package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the migration directory structure",
	Long:  `Creates the recommended directory structure for Casbin policy migrations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Define structure
		directories := []string{
			"casbin/policy_migrations",
		}

		for _, dir := range directories {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			fmt.Printf("Created directory: %s\n", dir)
		}

		// Create a sample model.conf if it doesn't exist
		modelPath := "casbin/model.conf"
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			content := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
			if err := os.WriteFile(modelPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create sample model.conf: %w", err)
			}
			fmt.Printf("Created sample file: %s\n", modelPath)
		}

		fmt.Println("Initialization complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
