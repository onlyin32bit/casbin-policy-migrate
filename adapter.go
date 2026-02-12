package casbinMigrate

import (
	"context"
)

// Adapter defines the interface for persisting migration history and applying policies.
// It allows for custom storage backends (e.g., pgx, gorm) or wrapping existing Casbin adapters.
type Adapter interface {
	// Migration History Management

	// Initialize sets up the necessary storage for migrations (e.g., creating the migrations table).
	Initialize(ctx context.Context) error

	// GetAppliedMigrations returns a list of all migrations that have been applied.
	GetAppliedMigrations(ctx context.Context) ([]Migration, error)

	// MarkMigrationApplied records that a migration has been successfully applied.
	MarkMigrationApplied(ctx context.Context, migration Migration) error

	// MarkMigrationRolledBack removes the record of a migration execution.
	MarkMigrationRolledBack(ctx context.Context, migration Migration) error

	// Policy Management
	// These methods are used to apply the changes defined in the migration files.
	// Implementing adapters should handle transactions if possible.

	// AddPolicy adds a policy rule. It should return success if the rule is added.
	AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error

	// RemovePolicy removes a policy rule. It should ignore if the rule does not exist.
	RemovePolicy(ctx context.Context, sec string, ptype string, rule []string) error
}
